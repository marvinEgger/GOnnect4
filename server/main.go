// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/marvinEgger/GOnnect4/server/lib"
)

const (
	initialClockDuration = 150 * time.Second // 2min 30s
	reconnectGracePeriod = 60 * time.Second
	defaultListenAddress = ":8080"
	webFolder            = "./client"
)

// Server manages all games and player connections
type Server struct {
	mu          sync.RWMutex
	gamesByCode map[string]*lib.Game
	lobby       map[lib.PlayerID]*lib.Player
}

// NewServer creates a new game server
func NewServer() *Server {
	return &Server{
		gamesByCode: make(map[string]*lib.Game),
		lobby:       make(map[lib.PlayerID]*lib.Player),
	}
}

// sendError sends an error message to a client
func (s *Server) sendError(client *lib.Client, message string) {
	client.Send(lib.Message{
		Type: lib.MsgError,
		Data: lib.ErrorData{Message: message},
	})
}

// handleLogin processes login/reconnection
func (s *Server) handleLogin(client *lib.Client, data lib.LoginData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var player *lib.Player
	var game *lib.Game

	// Reconnection attempt
	if data.PlayerID != nil {
		if p, exists := s.lobby[*data.PlayerID]; exists {
			player = p
			player.Username = data.Username

			// Check if player is in a game
			for code, g := range s.gamesByCode {
				if g.HasPlayer(player.ID) {
					game = g
					client.GameCode = code
					break
				}
			}
		}
	}
	// New player
	if player == nil {
		player = lib.NewPlayer(data.Username, initialClockDuration)
		if player == nil {
			s.sendError(client, "Invalid username")
			return
		}
		s.lobby[player.ID] = player
	}

	// Associate client with player
	client.PlayerID = player.ID
	player.SetSender(client)

	// Send welcome
	player.Send(lib.Message{
		Type: lib.MsgWelcome,
		Data: lib.WelcomeData{
			PlayerID: player.ID,
			Username: player.Username,
		},
	})

	// If reconnecting to a game, send game state
	if game != nil {
		s.sendGameState(player, game)
	}
}

// handleCreateGame creates a new game
func (s *Server) handleCreateGame(client *lib.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	player := s.lobby[client.PlayerID]
	if player == nil {
		s.sendError(client, "Player not found")
		return
	}

	game := lib.NewGame(initialClockDuration)
	game.TimerCallback = s.handleTimeout
	game.AddPlayer(player)
	s.gamesByCode[game.Code] = game
	client.GameCode = game.Code

	player.Send(lib.Message{
		Type: lib.MsgGameCreated,
		Data: lib.GameCreatedData{Code: game.Code},
	})

	s.sendGameState(player, game)
}

// handleJoinGame joins an existing game
func (s *Server) handleJoinGame(client *lib.Client, data lib.JoinGameData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.gamesByCode[data.Code]
	if !exists {
		s.sendError(client, "Game not found")
		return
	}

	player := s.lobby[client.PlayerID]
	if player == nil {
		s.sendError(client, "Player not found")
		return
	}

	// Reconnection
	if game.HasPlayer(player.ID) {
		client.GameCode = game.Code
		s.sendGameState(player, game)
		s.broadcastToGame(game, lib.Message{
			Type: lib.MsgGameState,
			Data: s.buildGameState(game, player.ID),
		})
		return
	}

	// New join
	if !game.AddPlayer(player) {
		s.sendError(client, "Cannot join game")
		return
	}

	client.GameCode = game.Code

	// Notify both players
	s.broadcastToGame(game, lib.Message{
		Type: lib.MsgGameStart,
		Data: lib.GameStartData{
			Code:          game.Code,
			CurrentTurn:   game.CurrentTurn,
			Players:       s.getPlayerInfos(game),
			TimeRemaining: s.getTimeRemaining(game),
		},
	})
}

// handlePlay processes a move
func (s *Server) handlePlay(client *lib.Client, data lib.PlayData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.gamesByCode[client.GameCode]
	if !exists {
		s.sendError(client, "Game not found")
		return
	}

	playerIdx := game.GetPlayerIndex(client.PlayerID)
	if playerIdx < 0 {
		s.sendError(client, "Player not in game")
		return
	}

	err := game.Play(playerIdx, data.Column)
	if err != nil {
		s.sendError(client, err.Error())
		return
	}

	// Broadcast move
	node := game.Board.GetLastPlayedNode(data.Column)
	s.broadcastToGame(game, lib.Message{
		Type: lib.MsgMove,
		Data: lib.MoveData{
			PlayerIdx:     playerIdx,
			Column:        data.Column,
			Row:           node.Row,
			Board:         game.Board.ToArray(),
			NextTurn:      game.CurrentTurn,
			TimeRemaining: s.getTimeRemaining(game),
		},
	})

	// Check game over
	if game.GetStatus() == lib.StatusFinished {
		s.broadcastToGame(game, lib.Message{
			Type: lib.MsgGameOver,
			Data: lib.GameOverData{
				Result: game.Result,
				Board:  game.Board.ToArray(),
			},
		})
	}
}

// handleReplay processes replay request
func (s *Server) handleReplay(client *lib.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.gamesByCode[client.GameCode]
	if !exists {
		s.sendError(client, "Game not found")
		return
	}

	playerIdx := game.GetPlayerIndex(client.PlayerID)
	if playerIdx < 0 {
		s.sendError(client, "Player not in game")
		return
	}

	// Broadcast replay request
	s.broadcastToGame(game, lib.Message{
		Type: lib.MsgReplayReq,
		Data: lib.ReplayRequestData{PlayerIdx: playerIdx},
	})

	// Check if both agreed
	if game.RequestReplay(playerIdx) {
		s.broadcastToGame(game, lib.Message{
			Type: lib.MsgGameStart,
			Data: lib.GameStartData{
				Code:          game.Code,
				CurrentTurn:   game.CurrentTurn,
				Players:       s.getPlayerInfos(game),
				TimeRemaining: s.getTimeRemaining(game),
			},
		})
	}
}

// handleForfeit processes forfeit request
func (s *Server) handleForfeit(client *lib.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.gamesByCode[client.GameCode]
	if !exists {
		s.sendError(client, "Game not found")
		return
	}

	playerIdx := game.GetPlayerIndex(client.PlayerID)
	if playerIdx < 0 {
		s.sendError(client, "Player not in game")
		return
	}

	game.Forfeit(playerIdx)

	s.broadcastToGame(game, lib.Message{
		Type: lib.MsgGameOver,
		Data: lib.GameOverData{
			Result: game.Result,
			Board:  game.Board.ToArray(),
		},
	})
}

// handleLeaveLobby processes leave lobby request
func (s *Server) handleLeaveLobby(client *lib.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If player is in a waiting game, delete it
	if client.GameCode != "" {
		if game, exists := s.gamesByCode[client.GameCode]; exists {
			if game.GetStatus() == lib.StatusWaiting {
				game.Cleanup()
				delete(s.gamesByCode, client.GameCode)
			} else if game.GetStatus() == lib.StatusPlaying {
				playerIdx := game.GetPlayerIndex(client.PlayerID)
				if playerIdx >= 0 {
					game.Forfeit(playerIdx)
					s.broadcastToGame(game, lib.Message{
						Type: lib.MsgGameOver,
						Data: lib.GameOverData{
							Result: game.Result,
							Board:  game.Board.ToArray(),
						},
					})
				}
			}
		}
	}

	client.GameCode = ""

	// Send welcome to return to lobby
	player := s.lobby[client.PlayerID]
	if player != nil {
		player.Send(lib.Message{
			Type: lib.MsgWelcome,
			Data: lib.WelcomeData{
				PlayerID: player.ID,
				Username: player.Username,
			},
		})
	}
}

// sendGameState sends current game state to a player
func (s *Server) sendGameState(player *lib.Player, game *lib.Game) {
	player.Send(lib.Message{
		Type: lib.MsgGameState,
		Data: s.buildGameState(game, player.ID),
	})
}

// buildGameState constructs game state data
func (s *Server) buildGameState(game *lib.Game, playerID lib.PlayerID) lib.GameStateData {
	return lib.GameStateData{
		Code:           game.Code,
		Status:         game.GetStatus(),
		Result:         game.Result,
		Board:          game.Board.ToArray(),
		Players:        s.getPlayerInfos(game),
		PlayerIdx:      game.GetPlayerIndex(playerID),
		CurrentTurn:    game.CurrentTurn,
		MoveCount:      game.MoveCount,
		TimeRemaining:  s.getTimeRemaining(game),
		ReplayRequests: game.ReplayRequests,
	}
}

// getPlayerInfos gets public info for both players
func (s *Server) getPlayerInfos(game *lib.Game) [2]lib.PlayerInfo {
	var infos [2]lib.PlayerInfo
	players := game.GetPlayers()
	for i, p := range players {
		if p != nil {
			infos[i] = lib.PlayerInfo{
				ID:        p.ID,
				Username:  p.Username,
				Connected: p.IsConnected(),
			}
		}
	}
	return infos
}

// getTimeRemaining gets remaining time for both players in milliseconds
func (s *Server) getTimeRemaining(game *lib.Game) [2]int64 {
	times := game.GetTimeRemaining()
	return [2]int64{
		times[0].Milliseconds(),
		times[1].Milliseconds(),
	}
}

// handleTimeout is called when a player's timer expires
func (s *Server) handleTimeout(gameCode string, loserIdx int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.gamesByCode[gameCode]
	if !exists || game.GetStatus() != lib.StatusPlaying {
		return
	}

	opponentIdx := 1 - loserIdx
	game.Status = lib.StatusFinished
	game.Result = lib.GameResult(opponentIdx + 1)

	s.broadcastToGame(game, lib.Message{
		Type: lib.MsgGameOver,
		Data: lib.GameOverData{
			Result: game.Result,
			Board:  game.Board.ToArray(),
		},
	})
}

// broadcastToGame sends a message to all players in a game
func (s *Server) broadcastToGame(game *lib.Game, msg lib.Message) {
	players := game.GetPlayers()
	for _, p := range players {
		if p != nil {
			p.Send(msg)
		}
	}
}

// cleanupStaleGames removes finished games and disconnected players
func (s *Server) cleanupStaleGames() {
	now := time.Now()

	for code, game := range s.gamesByCode {
		shouldDelete := false

		if game.GetStatus() == lib.StatusFinished {
			if now.Sub(game.LastPlayedAt) > reconnectGracePeriod {
				shouldDelete = true
			}
		} else if game.GetStatus() == lib.StatusWaiting {
			players := game.GetPlayers()
			if players[0] == nil || now.Sub(game.CreatedAt) > reconnectGracePeriod {
				shouldDelete = true
			}
		} else if game.GetStatus() == lib.StatusPlaying {
			bothDisconnected := true
			players := game.GetPlayers()
			for _, p := range players {
				if p != nil && p.IsConnected() {
					bothDisconnected = false
					break
				}
			}
			if bothDisconnected && now.Sub(game.LastPlayedAt) > reconnectGracePeriod {
				shouldDelete = true
			}
		}

		if shouldDelete {
			game.Cleanup()
			delete(s.gamesByCode, code)
		}
	}

	for id, player := range s.lobby {
		if !player.IsConnected() {
			inGame := false
			for _, game := range s.gamesByCode {
				if game.HasPlayer(id) {
					inGame = true
					break
				}
			}
			if !inGame {
				delete(s.lobby, id)
			}
		}
	}
}

// handleWebSocket handles websocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Printf("Failed to accept websocket: %v", err)
		return
	}

	client := lib.NewClient(conn)

	go client.WritePump()

	defer func() {
		s.mu.Lock()
		if client.PlayerID != "" {
			if player := s.lobby[client.PlayerID]; player != nil {
				// Only unset if this client is still the active sender
				// This handles race conditions where a new connection might have taken over
				// We can't easily check equality of interfaces, but we can check if it's connected
				// A more robust way would be to have an ID on the sender, but for now:
				player.SetSender(nil)
			}
		}
		s.cleanupStaleGames()
		s.mu.Unlock()
		client.Close()
	}()

	ctx := r.Context()
	for {
		var msg lib.Message
		err := wsjson.Read(ctx, conn, &msg)
		if err != nil {
			break
		}

		s.handleMessage(client, msg)
	}
}

// handleMessage routes messages to appropriate handlers
func (s *Server) handleMessage(client *lib.Client, msg lib.Message) {
	switch msg.Type {
	case lib.MsgLogin:
		var data lib.LoginData
		if err := mapToStruct(msg.Data, &data); err == nil {
			s.handleLogin(client, data)
		}

	case lib.MsgCreateGame:
		s.handleCreateGame(client)

	case lib.MsgJoinGame:
		var data lib.JoinGameData
		if err := mapToStruct(msg.Data, &data); err == nil {
			s.handleJoinGame(client, data)
		}

	case lib.MsgPlay:
		var data lib.PlayData
		if err := mapToStruct(msg.Data, &data); err == nil {
			s.handlePlay(client, data)
		}

	case lib.MsgReplay:
		s.handleReplay(client)

	case lib.MsgForfeit:
		s.handleForfeit(client)

	case lib.MsgLeaveLobby:
		s.handleLeaveLobby(client)
	}
}

// mapToStruct converts interface{} to struct via JSON
func mapToStruct(in interface{}, out interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func main() {
	server := NewServer()

	http.HandleFunc("/ws", server.handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir(webFolder)))

	addr := defaultListenAddress
	fmt.Printf("Server starting on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

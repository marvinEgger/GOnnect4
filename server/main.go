// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package main

import (
	"context"
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
	mu              sync.RWMutex
	gamesByCode     map[string]*lib.Game
	lobby           map[lib.PlayerID]*lib.Player
	bindingBySocket map[*websocket.Conn]*Binding
}

// Binding links a websocket connection to a player and game
type Binding struct {
	GameCode string
	PlayerID lib.PlayerID
}

// NewServer  creates a new game server
func NewServer() *Server {
	return &Server{
		gamesByCode:     make(map[string]*lib.Game),
		lobby:           make(map[lib.PlayerID]*lib.Player),
		bindingBySocket: make(map[*websocket.Conn]*Binding),
	}
}

// writeJSON sends a JSON message to a websocket
func writeJSON(ctx context.Context, conn *websocket.Conn, msg lib.Message) error {
	return wsjson.Write(ctx, conn, msg)
}

// sendError sends an error message to a client
func (s *Server) sendError(ctx context.Context, conn *websocket.Conn, message string) {
	_ = writeJSON(ctx, conn, lib.Message{
		Type: lib.MsgError,
		Data: lib.ErrorData{Message: message},
	})
}

// handleLogin processes login/reconnection
func (s *Server) handleLogin(ctx context.Context, conn *websocket.Conn, binding *Binding, data lib.LoginData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var player *lib.Player
	var game *lib.Game

	// Reconnection attempt
	if data.PlayerID != nil {
		if p, exists := s.lobby[*data.PlayerID]; exists {
			player = p
			player.SetConnected(true)
			player.Username = data.Username

			// Check if player is in a game
			for code, g := range s.gamesByCode {
				if g.HasPlayer(player.ID) {
					game = g
					binding.GameCode = code
					break
				}
			}
		}
	}
	// New player
	if player == nil {
		player = lib.NewPlayer(data.Username, initialClockDuration)
		if player == nil {
			s.sendError(ctx, conn, "Invalid username")
			return
		}
		s.lobby[player.ID] = player
	}

	binding.PlayerID = player.ID

	// Send welcome
	_ = writeJSON(ctx, conn, lib.Message{
		Type: lib.MsgWelcome,
		Data: lib.WelcomeData{
			PlayerID: player.ID,
			Username: player.Username,
		},
	})

	// If reconnecting to a game, send game state
	if game != nil {
		s.sendGameState(ctx, conn, game, player.ID)
	}
}

// handleCreateGame creates a new game
func (s *Server) handleCreateGame(ctx context.Context, conn *websocket.Conn, binding *Binding) {
	s.mu.Lock()
	defer s.mu.Unlock()

	player := s.lobby[binding.PlayerID]
	if player == nil {
		s.sendError(ctx, conn, "Player not found")
		return
	}

	game := lib.NewGame(initialClockDuration)
	game.TimerCallback = s.handleTimeout
	game.AddPlayer(player)
	s.gamesByCode[game.Code] = game
	binding.GameCode = game.Code

	_ = writeJSON(ctx, conn, lib.Message{
		Type: lib.MsgGameCreated,
		Data: lib.GameCreatedData{Code: game.Code},
	})

	s.sendGameState(ctx, conn, game, binding.PlayerID)

}

// handleJoinGame joins an existing game
func (s *Server) handleJoinGame(ctx context.Context, conn *websocket.Conn, binding *Binding, data lib.JoinGameData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.gamesByCode[data.Code]
	if !exists {
		s.sendError(ctx, conn, "Game not found")
		return
	}

	player := s.lobby[binding.PlayerID]
	if player == nil {
		s.sendError(ctx, conn, "Player not found")
		return
	}

	// Reconnection
	if game.HasPlayer(player.ID) {
		binding.GameCode = game.Code
		s.sendGameState(ctx, conn, game, player.ID)
		s.broadcastToGame(game, lib.Message{
			Type: lib.MsgGameState,
			Data: s.buildGameState(game, player.ID),
		})
		return
	}

	// New join
	if !game.AddPlayer(player) {
		s.sendError(ctx, conn, "Cannot join game")
		return
	}

	binding.GameCode = game.Code

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
func (s *Server) handlePlay(ctx context.Context, conn *websocket.Conn, binding *Binding, data lib.PlayData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.gamesByCode[binding.GameCode]
	if !exists {
		s.sendError(ctx, conn, "Game not found")
		return
	}

	playerIdx := game.GetPlayerIndex(binding.PlayerID)
	if playerIdx < 0 {
		s.sendError(ctx, conn, "Player not in game")
		return
	}

	err := game.Play(playerIdx, data.Column)
	if err != nil {
		s.sendError(ctx, conn, err.Error())
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
	if game.Status == lib.StatusFinished {
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
func (s *Server) handleReplay(ctx context.Context, conn *websocket.Conn, binding *Binding) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.gamesByCode[binding.GameCode]
	if !exists {
		s.sendError(ctx, conn, "Game not found")
		return
	}

	playerIdx := game.GetPlayerIndex(binding.PlayerID)
	if playerIdx < 0 {
		s.sendError(ctx, conn, "Player not in game")
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
func (s *Server) handleForfeit(ctx context.Context, conn *websocket.Conn, binding *Binding) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.gamesByCode[binding.GameCode]
	if !exists {
		s.sendError(ctx, conn, "Game not found")
		return
	}

	playerIdx := game.GetPlayerIndex(binding.PlayerID)
	if playerIdx < 0 {
		s.sendError(ctx, conn, "Player not in game")
		return
	}

	// Opponent wins
	opponentIdx := 1 - playerIdx
	game.Status = lib.StatusFinished
	game.Result = lib.GameResult(opponentIdx + 1) // 1 for player 0, 2 for player 1

	s.broadcastToGame(game, lib.Message{
		Type: lib.MsgGameOver,
		Data: lib.GameOverData{
			Result: game.Result,
			Board:  game.Board.ToArray(),
		},
	})
}

// handleLeaveLobby processes leave lobby request
func (s *Server) handleLeaveLobby(ctx context.Context, conn *websocket.Conn, binding *Binding) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If player is in a waiting game, delete it
	if binding.GameCode != "" {
		if game, exists := s.gamesByCode[binding.GameCode]; exists {
			if game.Status == lib.StatusWaiting {
				// Clean up the waiting game
				game.Cleanup()
				delete(s.gamesByCode, binding.GameCode)
			}
		}
	}

	// Clear game binding
	binding.GameCode = ""

	// Send welcome to return to lobby
	player := s.lobby[binding.PlayerID]
	if player != nil {
		_ = writeJSON(ctx, conn, lib.Message{
			Type: lib.MsgWelcome,
			Data: lib.WelcomeData{
				PlayerID: player.ID,
				Username: player.Username,
			},
		})
	}

}

// sendGameState sends current game state to a player
func (s *Server) sendGameState(ctx context.Context, conn *websocket.Conn, game *lib.Game, playerID lib.PlayerID) {
	_ = writeJSON(ctx, conn, lib.Message{
		Type: lib.MsgGameState,
		Data: s.buildGameState(game, playerID),
	})
}

// buildGameState constructs game state data
func (s *Server) buildGameState(game *lib.Game, playerID lib.PlayerID) lib.GameStateData {
	return lib.GameStateData{
		Code:          game.Code,
		Status:        game.Status,
		Result:        game.Result,
		Board:         game.Board.ToArray(),
		Players:       s.getPlayerInfos(game),
		PlayerIdx:     game.GetPlayerIndex(playerID),
		CurrentTurn:   game.CurrentTurn,
		MoveCount:     game.MoveCount,
		TimeRemaining: s.getTimeRemaining(game),
	}
}

// getPlayerInfos gets public info for both players
func (s *Server) getPlayerInfos(game *lib.Game) [2]lib.PlayerInfo {
	var infos [2]lib.PlayerInfo
	for i, p := range game.Players {
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
	if !exists || game.Status != lib.StatusPlaying {
		return
	}

	// Opponent wins
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
	for conn, binding := range s.bindingBySocket {
		if binding.GameCode == game.Code {
			_ = writeJSON(context.Background(), conn, msg)
		}
	}
}

// cleanupStaleGames removes finished games and disconnected players
func (s *Server) cleanupStaleGames() {
	now := time.Now()

	// Remove finished games older than grace period
	for code, game := range s.gamesByCode {
		shouldDelete := false

		if game.Status == lib.StatusFinished {
			// Delete finished games after grace period
			if now.Sub(game.LastPlayedAt) > reconnectGracePeriod {
				shouldDelete = true
			}
		} else if game.Status == lib.StatusWaiting {
			// Delete waiting games with no players or old games
			if game.Players[0] == nil || now.Sub(game.CreatedAt) > reconnectGracePeriod {
				shouldDelete = true
			}
		} else if game.Status == lib.StatusPlaying {
			// Check if both players disconnected
			bothDisconnected := true
			for _, p := range game.Players {
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

	// Remove disconnected players not in any game
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
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx := r.Context()
	binding := &Binding{}

	s.mu.Lock()
	s.bindingBySocket[conn] = binding
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.bindingBySocket, conn)
		if binding.PlayerID != "" {
			if player := s.lobby[binding.PlayerID]; player != nil {
				player.SetConnected(false)
			}
		}
		s.cleanupStaleGames()
		s.mu.Unlock()
	}()

	for {
		var msg lib.Message
		err := wsjson.Read(ctx, conn, &msg)
		if err != nil {
			break
		}

		s.handleMessage(ctx, conn, binding, msg)
	}

}

// handleMessage routes messages to appropriate handlers
func (s *Server) handleMessage(ctx context.Context, conn *websocket.Conn, binding *Binding, msg lib.Message) {
	switch msg.Type {
	case lib.MsgLogin:
		var data lib.LoginData
		if err := mapToStruct(msg.Data, &data); err == nil {
			s.handleLogin(ctx, conn, binding, data)
		}

	case lib.MsgCreateGame:
		s.handleCreateGame(ctx, conn, binding)

	case lib.MsgJoinGame:
		var data lib.JoinGameData
		if err := mapToStruct(msg.Data, &data); err == nil {
			s.handleJoinGame(ctx, conn, binding, data)
		}

	case lib.MsgPlay:
		var data lib.PlayData
		if err := mapToStruct(msg.Data, &data); err == nil {
			s.handlePlay(ctx, conn, binding, data)
		}

	case lib.MsgReplay:
		s.handleReplay(ctx, conn, binding)

	case lib.MsgForfeit:
		s.handleForfeit(ctx, conn, binding)

	case lib.MsgLeaveLobby:
		s.handleLeaveLobby(ctx, conn, binding)
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

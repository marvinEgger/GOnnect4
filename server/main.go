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
	"strings"
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
	mu               sync.RWMutex
	gamesByCode      map[string]*lib.Game
	lobby            map[lib.PlayerID]*lib.Player
	matchmakingQueue []lib.PlayerID // queue of players waiting for matchmaking

	// Background cleanup
	ctx        context.Context
	cancelFunc context.CancelFunc

	// Queue update throttling
	queueUpdatePending bool
	queueUpdateTimer   *time.Timer
}

// NewServer creates a new game server
func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		gamesByCode:      make(map[string]*lib.Game),
		lobby:            make(map[lib.PlayerID]*lib.Player),
		matchmakingQueue: make([]lib.PlayerID, 0),
		ctx:              ctx,
		cancelFunc:       cancel,
	}
}

// StartPeriodicCleanup starts a background goroutine that cleans up stale games every 30 seconds
func (s *Server) StartPeriodicCleanup() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				s.cleanupStaleGames()
				s.mu.Unlock()
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

// sendError sends an error message to a client
func (s *Server) sendError(client *lib.Client, message string) {
	client.Send(lib.Message{
		Type: lib.MsgError,
		Data: lib.ErrorData{Message: message},
	})
}

// findGameForClient finds and caches the game for a client
func (s *Server) findGameForClient(client *lib.Client) *lib.Game {
	// Try cached game code first
	if client.GameCode != "" {
		if game, exists := s.gamesByCode[client.GameCode]; exists {
			return game
		}
	}

	// Search all games for this player
	for code, game := range s.gamesByCode {
		if game.HasPlayer(client.PlayerID) {
			client.GameCode = code // Cache it
			return game
		}
	}

	return nil
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

	// Keep only first 5 characters (ASCII assumption)
	if len(data.Code) > 5 {
		data.Code = data.Code[:5]
	}

	// Uppercase
	data.Code = strings.ToUpper(data.Code)

	// Check if game exists
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

	game := s.findGameForClient(client)
	if game == nil {
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

	game := s.findGameForClient(client)
	if game == nil {
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

	game := s.findGameForClient(client)
	if game == nil {
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

// handleJoinMatchmaking adds player to matchmaking queue
func (s *Server) handleJoinMatchmaking(client *lib.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	player := s.lobby[client.PlayerID]
	if player == nil {
		s.sendError(client, "Player not found")
		return
	}

	// Check if already in queue
	for _, pid := range s.matchmakingQueue {
		if pid == client.PlayerID {
			return // Already in queue
		}
	}

	// Add to queue
	s.matchmakingQueue = append(s.matchmakingQueue, client.PlayerID)

	// Send searching confirmation
	player.Send(lib.Message{
		Type: lib.MsgMatchmakingSearching,
		Data: nil,
	})

	// Broadcast queue update to all players in lobby
	s.broadcastQueueUpdate()

	// Try to match players immediately if we have enough
	if len(s.matchmakingQueue) >= 2 {
		s.tryMatchPlayers()
	}
}

// handleLeaveMatchmaking removes player from matchmaking queue
func (s *Server) handleLeaveMatchmaking(client *lib.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from queue
	for i, pid := range s.matchmakingQueue {
		if pid == client.PlayerID {
			s.matchmakingQueue = append(s.matchmakingQueue[:i], s.matchmakingQueue[i+1:]...)
			break
		}
	}

	// Broadcast queue update
	s.broadcastQueueUpdate()
}

// tryMatchPlayers attempts to match two players in the queue
func (s *Server) tryMatchPlayers() {
	if len(s.matchmakingQueue) < 2 {
		return
	}

	// Take first two players
	player1ID := s.matchmakingQueue[0]
	player2ID := s.matchmakingQueue[1]

	// Remove from queue
	s.matchmakingQueue = s.matchmakingQueue[2:]

	player1 := s.lobby[player1ID]
	player2 := s.lobby[player2ID]

	// Verify both players still exist and are connected
	if player1 == nil || player2 == nil || !player1.IsConnected() || !player2.IsConnected() {
		// If one is missing, put the other back in queue
		if player1 != nil && player1.IsConnected() {
			s.matchmakingQueue = append([]lib.PlayerID{player1ID}, s.matchmakingQueue...)
		}
		if player2 != nil && player2.IsConnected() {
			s.matchmakingQueue = append([]lib.PlayerID{player2ID}, s.matchmakingQueue...)
		}
		s.broadcastQueueUpdate()
		return
	}

	// Create game
	game := lib.NewGame(initialClockDuration)
	game.TimerCallback = s.handleTimeout
	game.AddPlayer(player1)
	game.AddPlayer(player2)
	s.gamesByCode[game.Code] = game

	// Broadcast queue update after matching
	s.broadcastQueueUpdate()

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

// broadcastQueueUpdate sends queue size to all connected players
func (s *Server) broadcastQueueUpdate() {
	// Throttle queue updates to max once per 500ms to reduce load
	if s.queueUpdatePending {
		return // Already scheduled
	}

	s.queueUpdatePending = true

	// Stop existing timer if any
	if s.queueUpdateTimer != nil {
		s.queueUpdateTimer.Stop()
	}

	// Schedule the actual broadcast after 500ms
	s.queueUpdateTimer = time.AfterFunc(500*time.Millisecond, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.queueUpdatePending = false
		queueSize := len(s.matchmakingQueue)
		msg := lib.Message{
			Type: lib.MsgQueueUpdate,
			Data: lib.QueueUpdateData{
				PlayersInQueue: queueSize,
			},
		}

		// Send to all connected players in lobby
		for _, player := range s.lobby {
			if player != nil && player.IsConnected() {
				player.Send(msg)
			}
		}
	})
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
		LastMove:       game.LastMove,
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
	if !exists {
		return
	}

	// Use thread-safe Forfeit method to avoid race conditions
	game.Forfeit(loserIdx)

	// Broadcast game over only if forfeit was successful
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

			// Remove from matchmaking queue if present
			wasInQueue := false
			for i, pid := range s.matchmakingQueue {
				if pid == client.PlayerID {
					s.matchmakingQueue = append(s.matchmakingQueue[:i], s.matchmakingQueue[i+1:]...)
					wasInQueue = true
					break
				}
			}

			// Broadcast queue update if player was removed
			if wasInQueue {
				s.broadcastQueueUpdate()
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

	case lib.MsgJoinMatchmaking:
		s.handleJoinMatchmaking(client)

	case lib.MsgLeaveMatchmaking:
		s.handleLeaveMatchmaking(client)
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
	server.StartPeriodicCleanup()

	http.HandleFunc("/ws", server.handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir(webFolder)))

	addr := defaultListenAddress
	fmt.Printf("Server starting on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

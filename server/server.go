// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author: Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package main

import (
	"context"
	"sync"
	"time"

	"github.com/marvinEgger/GOnnect4/server/lib"
)

const (
	initialClockDuration = 150 * time.Second // 2min 30s
	reconnectGracePeriod = 120 * time.Second
	cleanupInterval      = 30 * time.Second
	queueUpdateDelay     = 500 * time.Millisecond
)

// Server manages all games and player connections
type Server struct {
	mu               sync.RWMutex
	gamesByCode      map[string]*lib.Game
	lobby            map[lib.PlayerID]*lib.Player
	matchmakingQueue []lib.PlayerID

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

// StartPeriodicCleanup starts a background goroutine that cleans up stale games
func (srv *Server) StartPeriodicCleanup() {
	// Create ticker that fires every 30 seconds
	ticker := time.NewTicker(cleanupInterval)

	// Start background goroutine (runs independently)
	go func() {
		// Stop ticker when goroutine exits to free resources
		defer ticker.Stop()

		for {
			// Select waits for one of these events
			select {
			case <-ticker.C:
				// Ticker fired that will run cleanup
				srv.mu.Lock()
				srv.cleanupStaleGames()
				srv.mu.Unlock()

			case <-srv.ctx.Done():
				// When server shutdown, exit go routine
				return
			}
		}
	}()
}

// sendError sends an error message to a client
func (srv *Server) sendError(client *lib.Client, err error) {
	client.Send(lib.Message{
		Type: lib.MsgError,
		Data: lib.ErrorData{Message: err.Error()},
	})
}

// findGameForClient finds and caches the game for a client
func (srv *Server) findGameForClient(client *lib.Client) *lib.Game {
	// Try cached game code first
	if client.GameCode != "" {
		if game, exists := srv.gamesByCode[client.GameCode]; exists {
			return game
		}
	}

	// Search all games for this player
	for code, game := range srv.gamesByCode {
		if game.HasPlayer(client.PlayerID) {
			client.GameCode = code
			return game
		}
	}

	return nil
}

// sendGameState sends current game state to a player
func (srv *Server) sendGameState(player *lib.Player, game *lib.Game) {
	player.Send(lib.Message{
		Type: lib.MsgGameState,
		Data: srv.buildGameState(game, player.ID),
	})
}

// buildGameState constructs game state data
func (srv *Server) buildGameState(game *lib.Game, playerID lib.PlayerID) lib.GameStateData {
	return lib.GameStateData{
		Code:           game.Code,
		Status:         game.GetStatus(),
		Result:         game.Result,
		Board:          game.Board.ToArray(),
		Players:        srv.getPlayerInfos(game),
		PlayerIdx:      game.GetPlayerIndex(playerID),
		CurrentTurn:    game.CurrentTurn,
		MoveCount:      game.MoveCount,
		TimeRemaining:  srv.getTimeRemaining(game),
		ReplayRequests: game.ReplayRequests,
		LastMove:       game.LastMove,
	}
}

// broadcastToGame sends a message to all players in a game
func (srv *Server) broadcastToGame(game *lib.Game, msg lib.Message) {
	players := game.GetPlayers()
	for _, p := range players {
		if p != nil {
			p.Send(msg)
		}
	}
}

// handleTimeout is called when a player's timer expires
func (srv *Server) handleTimeout(gameCode string, loserIdx int) {
	// Lock needed because timer callback runs in separate goroutine
	srv.mu.Lock()
	defer srv.mu.Unlock()

	// Find game
	game, exists := srv.gamesByCode[gameCode]

	// Check if game exist, might have been deleted already
	if !exists {
		return
	}

	// Player loses by timeout
	game.Forfeit(loserIdx)

	// Notify both players if game actually ended
	if game.GetStatus() == lib.StatusFinished {
		srv.broadcastToGame(game, lib.Message{
			Type: lib.MsgGameOver,
			Data: lib.GameOverData{
				Result: game.Result,
				Board:  game.Board.ToArray(),
			},
		})
	}
}

// cleanupStaleGames removes finished games and disconnected players
func (srv *Server) cleanupStaleGames() {
	now := time.Now()

	// Clean up old games
	for code, game := range srv.gamesByCode {
		shouldDelete := false

		// If finished games, keep alive for a time to allow reconnection, then delete
		if game.GetStatus() == lib.StatusFinished {
			if now.Sub(game.LastPlayedAt) > reconnectGracePeriod {
				shouldDelete = true
			}

			// For waiting games delete if creator left or waited too long
		} else if game.GetStatus() == lib.StatusWaiting {
			players := game.GetPlayers()
			if players[0] == nil || now.Sub(game.CreatedAt) > reconnectGracePeriod {
				shouldDelete = true
			}

			// In active games delete only if both players disconnected for too long
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
			// Stop timers and free resources
			game.Cleanup()
			delete(srv.gamesByCode, code)
		}
	}

	// Clean up disconnected players not in any game
	for id, player := range srv.lobby {
		if !player.IsConnected() {
			// Check if player is in a game (allows reconnection)
			inGame := false
			for _, game := range srv.gamesByCode {
				if game.HasPlayer(id) {
					inGame = true
					break
				}
			}

			// Remove only if not in any game
			if !inGame {
				delete(srv.lobby, id)
			}
		}
	}
}

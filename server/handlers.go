// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author: Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package main

import (
	"strings"

	"github.com/marvinEgger/GOnnect4/server/lib"
)

const maxGameCodeLength = 5

// handleLogin processes login / reconnection
func (srv *Server) handleLogin(client *lib.Client, data lib.LoginData) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	var player *lib.Player
	var game *lib.Game

	// Reconnection attempt
	if data.PlayerID != nil {
		if p, exists := srv.lobby[*data.PlayerID]; exists {
			player = p
			player.Username = data.Username

			// Check if player is in a game
			for code, g := range srv.gamesByCode {
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
			srv.sendError(client, lib.ErrInvalidUsername)
			return
		}
		srv.lobby[player.ID] = player
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
		srv.sendGameState(player, game)
	}
}

// handleCreateGame creates a new game
func (srv *Server) handleCreateGame(client *lib.Client) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	// Verify player exists in lobby
	player := srv.lobby[client.PlayerID]
	if player == nil {
		srv.sendError(client, lib.ErrPlayerNotFound)
		return
	}

	// Check if player is already in an active game
	if client.GameCode != "" {
		if game, exists := srv.gamesByCode[client.GameCode]; exists {
			// Only allow creating new game if current game is finished
			if game.GetStatus() != lib.StatusFinished {
				srv.sendError(client, lib.ErrPlayerAlreadyInGame)
				return
			}
		}
	}

	// Create new game and add player as host
	game := lib.NewGame(initialClockDuration)
	game.TimerCallback = srv.handleTimeout
	game.AddPlayer(player)
	srv.gamesByCode[game.Code] = game
	client.GameCode = game.Code

	// Notify player of game creation
	player.Send(lib.Message{
		Type: lib.MsgGameCreated,
		Data: lib.GameCreatedData{Code: game.Code},
	})

	srv.sendGameState(player, game)
}

// handleJoinGame joins an existing game
func (srv *Server) handleJoinGame(client *lib.Client, data lib.JoinGameData) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	// Normalize game code (trim to 5 chars and uppercase)
	if len(data.Code) > maxGameCodeLength {
		data.Code = data.Code[:maxGameCodeLength]
	}
	data.Code = strings.ToUpper(data.Code)

	// Verify game exists
	game, exists := srv.gamesByCode[data.Code]
	if !exists {
		srv.sendError(client, lib.ErrGameNotFound)
		return
	}

	// Verify player exists in lobby
	player := srv.lobby[client.PlayerID]
	if player == nil {
		srv.sendError(client, lib.ErrPlayerNotFound)
		return
	}

	// Handle reconnection (player already in this game)
	if game.HasPlayer(player.ID) {
		client.GameCode = game.Code
		srv.sendGameState(player, game)
		srv.broadcastToGame(game, lib.Message{
			Type: lib.MsgGameState,
			Data: srv.buildGameState(game, player.ID),
		})
		return
	}

	// Add player to game (fails if game is full)
	if !game.AddPlayer(player) {
		srv.sendError(client, lib.ErrGameFull)
		return
	}

	client.GameCode = game.Code

	// Notify both players that game is starting
	srv.broadcastToGame(game, lib.Message{
		Type: lib.MsgGameStart,
		Data: lib.GameStartData{
			Code:          game.Code,
			CurrentTurn:   game.CurrentTurn,
			Players:       srv.getPlayerInfos(game),
			TimeRemaining: srv.getTimeRemaining(game),
		},
	})
}

// handlePlay processes a move
func (srv *Server) handlePlay(client *lib.Client, data lib.PlayData) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	game := srv.findGameForClient(client)
	if game == nil {
		srv.sendError(client, lib.ErrGameNotFound)
		return
	}

	playerIdx := game.GetPlayerIndex(client.PlayerID)
	if playerIdx < 0 {
		srv.sendError(client, lib.ErrPlayerNotInGame)
		return
	}

	err := game.Play(playerIdx, data.Column)
	if err != nil {
		srv.sendError(client, err)
		return
	}

	// Broadcast move
	node := game.Board.GetLastPlayedNode(data.Column)
	srv.broadcastToGame(game, lib.Message{
		Type: lib.MsgMove,
		Data: lib.MoveData{
			PlayerIdx:     playerIdx,
			Column:        data.Column,
			Row:           node.Row,
			Board:         game.Board.ToArray(),
			NextTurn:      game.CurrentTurn,
			TimeRemaining: srv.getTimeRemaining(game),
		},
	})

	// Check game over
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

// handleReplay processes replay request
func (srv *Server) handleReplay(client *lib.Client) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	game := srv.findGameForClient(client)
	if game == nil {
		srv.sendError(client, lib.ErrGameNotFound)
		return
	}

	playerIdx := game.GetPlayerIndex(client.PlayerID)
	if playerIdx < 0 {
		srv.sendError(client, lib.ErrPlayerNotInGame)
		return
	}

	// Broadcast replay request
	srv.broadcastToGame(game, lib.Message{
		Type: lib.MsgReplayReq,
		Data: lib.ReplayRequestData{PlayerIdx: playerIdx},
	})

	// Check if both agreed
	if game.RequestReplay(playerIdx) {
		srv.broadcastToGame(game, lib.Message{
			Type: lib.MsgGameStart,
			Data: lib.GameStartData{
				Code:          game.Code,
				CurrentTurn:   game.CurrentTurn,
				Players:       srv.getPlayerInfos(game),
				TimeRemaining: srv.getTimeRemaining(game),
			},
		})
	}
}

// handleForfeit processes forfeit request
func (srv *Server) handleForfeit(client *lib.Client) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	game := srv.findGameForClient(client)
	if game == nil {
		srv.sendError(client, lib.ErrGameNotFound)
		return
	}

	playerIdx := game.GetPlayerIndex(client.PlayerID)
	if playerIdx < 0 {
		srv.sendError(client, lib.ErrPlayerNotInGame)
		return
	}

	game.Forfeit(playerIdx)

	srv.broadcastToGame(game, lib.Message{
		Type: lib.MsgGameOver,
		Data: lib.GameOverData{
			Result: game.Result,
			Board:  game.Board.ToArray(),
		},
	})
}

// handleLeaveLobby processes leave lobby request
func (srv *Server) handleLeaveLobby(client *lib.Client) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	// Clean up player's current game if any
	if client.GameCode != "" {
		if game, exists := srv.gamesByCode[client.GameCode]; exists {
			// Waiting game: delete it (player was alone waiting for opponent)
			if game.GetStatus() == lib.StatusWaiting {
				game.Cleanup()
				delete(srv.gamesByCode, client.GameCode)
				// Active game: forfeit (opponent wins)
			} else if game.GetStatus() == lib.StatusPlaying {
				playerIdx := game.GetPlayerIndex(client.PlayerID)
				if playerIdx >= 0 {
					game.Forfeit(playerIdx)
					srv.broadcastToGame(game, lib.Message{
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

	// Clear player's game code
	client.GameCode = ""

	// Send welcome message to return player to lobby
	player := srv.lobby[client.PlayerID]
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

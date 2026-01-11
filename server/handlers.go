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

			// Clean up any finished games player is still in from previous session
			srv.cleanupPlayerFinishedGames(player.ID)

			// Check if player is in an ACTIVE game (waiting or playing only)
			for code, g := range srv.gamesByCode {
				if g.HasPlayer(player.ID) {
					status := g.GetStatus()
					if status == lib.StatusWaiting || status == lib.StatusPlaying {
						game = g
						client.SetGameCode(code)
						break
					}
				}
			}
		} else {
			// Reconnection failed - session not found
			srv.sendError(client, lib.ErrReconnectionFailed)
			return
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

	// If reconnecting to an active game, send game state
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
	if player == nil || !player.IsConnected() {
		srv.sendError(client, lib.ErrPlayerNotFound)
		return
	}

	// Clean up any finished games player is still in
	srv.cleanupPlayerFinishedGames(client.PlayerID)

	// Check if player is already in an active game (scan all games)
	if activeGame := srv.findActiveGameForPlayer(client.PlayerID); activeGame != nil {
		srv.sendError(client, lib.ErrPlayerAlreadyInGame)
		return
	}

	// Create new game and add player as host
	game := lib.NewGame(initialClockDuration)
	game.TimerCallback = srv.handleTimeout
	game.AddPlayer(player)
	srv.gamesByCode[game.Code] = game
	client.SetGameCode(game.Code)

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

	// Clean up any finished games player is still in (except this one)
	srv.cleanupPlayerFinishedGames(client.PlayerID)

	// Handle reconnection (player already in this game)
	if game.HasPlayer(player.ID) {
		client.SetGameCode(game.Code)
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

	client.SetGameCode(game.Code)

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
	if node == nil {
		srv.sendError(client, lib.ErrInvalidMove)
		return
	}

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

	// Handle waiting games differently
	if game.GetStatus() == lib.StatusWaiting {
		srv.deleteGame(client.GetGameCode())
		client.SetGameCode("")
		return
	}

	// Normal forfeit for active games
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
	if client.GetGameCode() != "" {
		if game, exists := srv.gamesByCode[client.GetGameCode()]; exists {
			// Waiting game: delete it (player was alone waiting for opponent)
			if game.GetStatus() == lib.StatusWaiting {
				// Active game: forfeit (opponent wins)
			srv.deleteGame(client.GetGameCode())
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
				// Finished game: player is leaving after game over
			} else if game.GetStatus() == lib.StatusFinished {
				// Check if both players have left
			playerIdx := game.GetPlayerIndex(client.PlayerID)

				if playerIdx >= 0 {
					// Remove player from game
					game.RemovePlayer(client.PlayerID)

					// If both players have left or only disconnected opponent remains, cleanup game
					players := game.GetPlayers()
					opponentIdx := 1 - playerIdx
					if players[opponentIdx] == nil || !players[opponentIdx].IsConnected() {
						srv.deleteGame(client.GetGameCode())
					}
				}
			}
		}
	}

	// Clear player's game code
	client.SetGameCode("")

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

// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author: Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package main

import (
	"time"

	"github.com/marvinEgger/GOnnect4/server/lib"
)

const minPlayersForMatch = 2

// handleJoinMatchmaking adds player to matchmaking queue
func (srv *Server) handleJoinMatchmaking(client *lib.Client) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	player := srv.lobby[client.PlayerID]
	if player == nil {
		srv.sendError(client, lib.ErrPlayerNotFound)
		return
	}

	// Check if already in queue
	for _, pid := range srv.matchmakingQueue {
		if pid == client.PlayerID {
			return
		}
	}

	// Add to queue
	srv.matchmakingQueue = append(srv.matchmakingQueue, client.PlayerID)

	// Send searching confirmation
	player.Send(lib.Message{
		Type: lib.MsgMatchmakingSearching,
		Data: nil,
	})

	// Broadcast queue update to all players in lobby
	srv.broadcastQueueUpdate()

	// Try to match players immediately if we have enough
	if len(srv.matchmakingQueue) >= minPlayersForMatch {
		srv.tryMatchPlayers()
	}
}

// handleLeaveMatchmaking removes player from matchmaking queue
func (srv *Server) handleLeaveMatchmaking(client *lib.Client) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	// Remove from queue
	for i, pid := range srv.matchmakingQueue {
		if pid == client.PlayerID {
			// Removes this player from the matchmaking lobby queue by rebuilding the queue slice without them (skip index i)
			srv.matchmakingQueue = append(srv.matchmakingQueue[:i], srv.matchmakingQueue[i+1:]...)
			break
		}
	}

	// Broadcast queue update
	srv.broadcastQueueUpdate()
}

// tryMatchPlayers attempts to match two players in the queue
func (srv *Server) tryMatchPlayers() {
	if len(srv.matchmakingQueue) < minPlayersForMatch {
		return
	}

	// Take first two players
	player1ID := srv.matchmakingQueue[0]
	player2ID := srv.matchmakingQueue[1]

	// Remove from queue
	srv.matchmakingQueue = srv.matchmakingQueue[2:]

	// Get players from lobby when it's their turn to play
	player1 := srv.lobby[player1ID]
	player2 := srv.lobby[player2ID]

	// Verify both players still exist and are connected
	if player1 == nil || player2 == nil || !player1.IsConnected() || !player2.IsConnected() {
		// If one is missing, put the other back in queue
		if player1 != nil && player1.IsConnected() {
			srv.matchmakingQueue = append([]lib.PlayerID{player1ID}, srv.matchmakingQueue...)
		}
		if player2 != nil && player2.IsConnected() {
			srv.matchmakingQueue = append([]lib.PlayerID{player2ID}, srv.matchmakingQueue...)
		}
		srv.broadcastQueueUpdate()
		return
	}

	// Create game
	game := lib.NewGame(initialClockDuration)
	game.TimerCallback = srv.handleTimeout
	game.AddPlayer(player1)
	game.AddPlayer(player2)
	srv.gamesByCode[game.Code] = game

	// Broadcast queue update after matching
	srv.broadcastQueueUpdate()

	// Notify both players
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

// broadcastQueueUpdate sends queue size to all connected players
func (srv *Server) broadcastQueueUpdate() {
	// Throttle queue updates to max once per 500ms to reduce load
	if srv.queueUpdatePending {
		return
	}

	srv.queueUpdatePending = true

	// Stop existing timer if any
	if srv.queueUpdateTimer != nil {
		srv.queueUpdateTimer.Stop()
	}

	// Schedule the actual broadcast after delay
	srv.queueUpdateTimer = time.AfterFunc(queueUpdateDelay, func() {
		srv.mu.Lock()
		defer srv.mu.Unlock()

		srv.queueUpdatePending = false
		queueSize := len(srv.matchmakingQueue)
		msg := lib.Message{
			Type: lib.MsgQueueUpdate,
			Data: lib.QueueUpdateData{
				PlayersInQueue: queueSize,
			},
		}

		// Send to all connected players in lobby
		for _, player := range srv.lobby {
			if player != nil && player.IsConnected() {
				player.Send(msg)
			}
		}
	})
}

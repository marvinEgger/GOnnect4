// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package main

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/marvinEgger/GOnnect4/server/lib"
)

// handleWebSocket handles websocket connections
func (srv *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Allow connections from any origin
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Printf("Failed to accept websocket: %v", err)
		return
	}

	// Create client wrapper for this connection
	client := lib.NewClient(conn)

	// Start write pump in separate goroutine to send messages to client
	// Runs concurrently to avoid blocking when sending multiple messages
	go client.WritePump()

	// Cleanup function runs when connection closes (error, disconnect, or normal exit)
	defer func() {
		srv.mu.Lock()
		if client.PlayerID != "" {
			// Disconnect player from lobby
			if player := srv.lobby[client.PlayerID]; player != nil {
				player.SetSender(nil)
			}

			// Remove from matchmaking queue if present
			wasInQueue := false
			for i, pid := range srv.matchmakingQueue {
				if pid == client.PlayerID {
					srv.matchmakingQueue = append(srv.matchmakingQueue[:i], srv.matchmakingQueue[i+1:]...)
					wasInQueue = true
					break
				}
			}

			// Notify other players in queue
			if wasInQueue {
				srv.broadcastQueueUpdate()
			}
		}
		// Clean up any stale games or disconnected players
		srv.cleanupStaleGames()
		srv.mu.Unlock()

		// Close WebSocket connection and stop write pump of client
		client.Close()
	}()

	// Get request context for cancellation handling
	ctx := r.Context()

	// Reading loop, continuously read messages from client until error or disconnect
	for {
		var msg lib.Message
		err := wsjson.Read(ctx, conn, &msg)
		if err != nil {
			// Connection closed or error, exit loop to run defer cleanup
			break
		}

		// Route message to appropriate handler
		srv.handleMessage(client, msg)
	}
}

// handleMessage routes messages to appropriate handlers
func (srv *Server) handleMessage(client *lib.Client, msg lib.Message) {
	switch msg.Type {
	case lib.MsgLogin:
		var data lib.LoginData
		if err := mapToStruct(msg.Data, &data); err == nil {
			srv.handleLogin(client, data)
		}

	case lib.MsgCreateGame:
		srv.handleCreateGame(client)

	case lib.MsgJoinGame:
		var data lib.JoinGameData
		if err := mapToStruct(msg.Data, &data); err == nil {
			srv.handleJoinGame(client, data)
		}

	case lib.MsgPlay:
		var data lib.PlayData
		if err := mapToStruct(msg.Data, &data); err == nil {
			srv.handlePlay(client, data)
		}

	case lib.MsgReplay:
		srv.handleReplay(client)

	case lib.MsgForfeit:
		srv.handleForfeit(client)

	case lib.MsgLeaveLobby:
		srv.handleLeaveLobby(client)

	case lib.MsgJoinMatchmaking:
		srv.handleJoinMatchmaking(client)

	case lib.MsgLeaveMatchmaking:
		srv.handleLeaveMatchmaking(client)
	}
}

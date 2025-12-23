// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author: Astrit Aslani astrit.aslani@gmail.com
// Created: 23.12.2025
//go:build js && wasm

package main

import (
	"sync"
	"syscall/js"
	"time"

	"github.com/marvinEgger/GOnnect4/client/wasm/lib"
)

const (
	autoConnectTimeout      = 3 * time.Second
	matchmakingMinDelay     = 500
	matchmakingMaxDelay     = 3000
	messageDisplayTime      = 3 * time.Second
	copyButtonResetTime     = 2 * time.Second
	errorMessageDisplayTime = 5 * time.Second
)

// setupEventListeners attaches all UI event listeners
func setupEventListeners() {
	// Login screen
	attachEventListener("connect-btn", "click", handleConnect)
	attachKeyPressListener("username-input", handleConnect)

	// Lobby - mode selection
	attachEventListener("friend-mode-btn", "click", handleFriendMode)
	attachEventListener("matchmaking-btn", "click", handleMatchmakingMode)
	attachEventListener("back-to-modes-btn", "click", handleBackToModes)
	attachEventListener("back-to-modes-matchmaking-btn", "click", handleBackToModes)
	attachEventListener("logout-btn-header", "click", handleLogout)

	// Friend mode
	attachEventListener("create-game-btn", "click", handleCreateGame)
	attachEventListener("join-game-btn", "click", handleJoinGame)
	attachKeyPressListener("join-code-input", handleJoinGame)
	attachEventListener("copy-code-btn", "click", handleCopyCode)

	// Matchmaking mode
	attachEventListener("cancel-matchmaking-btn", "click", handleCancelMatchmaking)

	// Game screen
	attachEventListener("copy-code-game-btn", "click", handleCopyGameCode)
	attachEventListener("replay-btn", "click", handleReplay)
	attachEventListener("forfeit-btn", "click", handleForfeit)
	attachEventListener("back-to-lobby-btn", "click", handleBackToLobby)
	attachEventListener("cancel-game-btn", "click", handleCancelGame)

	// Board interactions
	setupBoardListeners()
}

// attachEventListener attaches a simple event listener
func attachEventListener(elementID, eventType string, handler func(js.Value, []js.Value) interface{}) {
	element := lib.GetElement(elementID)
	if !element.IsNull() {
		element.Call("addEventListener", eventType, js.FuncOf(handler))
	}
}

// attachKeyPressListener attaches an Enter key listener
func attachKeyPressListener(elementID string, handler func(js.Value, []js.Value) interface{}) {
	element := lib.GetElement(elementID)
	if !element.IsNull() {
		element.Call("addEventListener", "keypress", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			event := args[0]
			if event.Get("key").String() == "Enter" {
				handler(this, args)
			}
			return nil
		}))
	}
}

// setupBoardListeners attaches canvas event listeners
func setupBoardListeners() {
	canvas := lib.GetElement("game-board")
	if canvas.IsNull() {
		return
	}

	canvas.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		lib.HandleClick(args[0])
		return nil
	}))

	canvas.Call("addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		lib.HandleHover(args[0])
		return nil
	}))

	canvas.Call("addEventListener", "mouseleave", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		lib.HandleLeave(args[0])
		return nil
	}))
}

// ============================================================================
// UI EVENT HANDLERS
// ============================================================================

// handleConnect processes user login
func handleConnect(this js.Value, args []js.Value) interface{} {
	username := lib.GetValue("username-input")
	if username == "" {
		lib.ShowMessage("login-message", "Please enter a username", "error")
		return nil
	}

	lib.Connect(username, "", handleMessage)
	return nil
}

// handleCreateGame creates a new private game
func handleCreateGame(this js.Value, args []js.Value) interface{} {
	lib.SendMessage("create_game", map[string]interface{}{})
	showWaitingArea()
	return nil
}

// handleJoinGame joins an existing game with code
func handleJoinGame(this js.Value, args []js.Value) interface{} {
	code := lib.GetValue("join-code-input")
	if code == "" {
		lib.ShowMessage("lobby-message", "Please enter a game code", "error")
		return nil
	}

	lib.SendMessage("join_game", map[string]interface{}{
		"code": code,
	})
	return nil
}

// handleCopyCode copies game code to clipboard
func handleCopyCode(this js.Value, args []js.Value) any {
	code := lib.Get().GetGameCode()

	js.Global().
		Get("navigator").
		Get("clipboard").
		Call("writeText", code)

	lib.ShowMessage("lobby-message", "Code copied!", "success")

	time.AfterFunc(messageDisplayTime, func() {
		clearMessage("lobby-message")
	})

	return js.Undefined()
}

// handleCopyGameCode copies game code to clipboard from game screen
func handleCopyGameCode(this js.Value, args []js.Value) any {
	code := lib.Get().GetGameCode()
	js.Global().Get("navigator").Get("clipboard").Call("writeText", code)

	lib.SetText("copy-code-game-btn", "Copied!")
	time.AfterFunc(copyButtonResetTime, func() {
		lib.SetText("copy-code-game-btn", "Copy")
	})

	return js.Undefined()
}

// handleReplay requests a game replay
func handleReplay(this js.Value, args []js.Value) interface{} {
	state := lib.Get()
	state.SetReplayRequested(true)
	lib.SendMessage("replay", map[string]interface{}{})
	updateReplayButton()
	return nil
}

// handleForfeit forfeits the current game
func handleForfeit(this js.Value, args []js.Value) interface{} {
	if lib.Confirm("Are you sure you want to forfeit? Your opponent will win.") {
		lib.SendMessage("forfeit", map[string]interface{}{})
	}
	return nil
}

// handleBackToLobby returns to lobby from game
func handleBackToLobby(this js.Value, args []js.Value) interface{} {
	lib.SendMessage("leave_lobby", map[string]interface{}{})
	return nil
}

// handleCancelGame cancels waiting for opponent
func handleCancelGame(this js.Value, args []js.Value) interface{} {
	lib.SendMessage("leave_lobby", map[string]interface{}{})
	return nil
}

// handleLogout logs out the current user
func handleLogout(this js.Value, args []js.Value) interface{} {
	lib.RemoveLocalStorage("playerID")
	lib.RemoveLocalStorage("username")

	lib.Close()
	lib.Stop()

	state := lib.Get()
	state.SetPlayerID("")
	state.SetGameCode("")
	state.SetPlayerIdx(-1)

	lib.SetValue("username-input", "")
	lib.Hide("header-user-info")
	lib.ShowScreen("login")

	return nil
}

// handleFriendMode shows friend mode panel
func handleFriendMode(this js.Value, args []js.Value) interface{} {
	lib.Hide("mode-selection")
	lib.Hide("matchmaking-panel")
	lib.Show("friend-mode-panel")
	return nil
}

// handleMatchmakingMode starts matchmaking
func handleMatchmakingMode(this js.Value, args []js.Value) interface{} {
	lib.Hide("mode-selection")
	lib.Hide("friend-mode-panel")
	lib.Show("matchmaking-panel")
	lib.Show("matchmaking-searching")

	// Add random delay to mix up players in queue
	randomDelay := matchmakingMinDelay + int(js.Global().Get("Math").Call("random").Float()*float64(matchmakingMaxDelay-matchmakingMinDelay))

	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		lib.SendMessage("join_matchmaking", map[string]interface{}{})
		return nil
	}), randomDelay)

	return nil
}

// handleBackToModes returns to mode selection
func handleBackToModes(this js.Value, args []js.Value) interface{} {
	lib.Hide("friend-mode-panel")
	lib.Hide("matchmaking-panel")
	lib.Hide("waiting-area")
	lib.Hide("matchmaking-searching")

	lib.SendMessage("leave_matchmaking", map[string]interface{}{})

	clearMessage("lobby-message")

	lib.Show("mode-selection")
	resetLobby()

	return nil
}

// handleCancelMatchmaking cancels matchmaking search
func handleCancelMatchmaking(this js.Value, args []js.Value) interface{} {
	lib.SendMessage("leave_matchmaking", map[string]interface{}{})

	lib.Hide("matchmaking-panel")
	lib.Show("mode-selection")

	return nil
}

// autoConnect attempts to reconnect with saved credentials
func autoConnect(username, savedPlayerID string) {
	done := make(chan struct{})
	timedOut := make(chan struct{})
	var once sync.Once

	// Timeout goroutine
	go func() {
		select {
		case <-time.After(autoConnectTimeout):
			close(timedOut)
			lib.Console("Auto-reconnect timeout - clearing saved credentials")
			lib.RemoveLocalStorage("playerID")
			lib.RemoveLocalStorage("username")
			lib.Close()
			lib.ShowScreen("login")
		case <-done:
			// Connected successfully before timeout
		}
	}()

	// Custom message handler that cancels timeout
	customHandler := func(msg lib.Message) {
		select {
		case <-timedOut:
			return
		default:
		}

		once.Do(func() { close(done) })
		handleMessage(msg)
	}

	lib.Connect(username, savedPlayerID, customHandler)
}

// ============================================================================
// WEBSOCKET MESSAGE HANDLERS
// ============================================================================

// handleMessage routes incoming WebSocket messages to appropriate handlers
func handleMessage(msg lib.Message) {
	switch msg.Type {
	case "welcome":
		handleWelcome(msg.Data)
	case "game_created":
		handleGameCreated(msg.Data)
	case "game_start":
		handleGameStart(msg.Data)
	case "game_state":
		handleGameState(msg.Data)
	case "move":
		handleMove(msg.Data)
	case "game_over":
		handleGameOver(msg.Data)
	case "replay_request":
		handleReplayRequest(msg.Data)
	case "matchmaking_searching":
		handleMatchmakingSearching(msg.Data)
	case "queue_update":
		handleQueueUpdate(msg.Data)
	case "error":
		handleError(msg.Data)
	}
}

// handleWelcome processes welcome message after login
func handleWelcome(data interface{}) {
	var welcome lib.WelcomeData
	if err := remarshal(data, &welcome); err != nil {
		lib.Console("handleWelcome: remarshal failed: " + err.Error())
		return
	}

	state := lib.Get()
	state.SetPlayerID(welcome.PlayerID)

	lib.SetLocalStorage("playerID", welcome.PlayerID)
	lib.SetLocalStorage("username", welcome.Username)

	lib.SetText("header-username", welcome.Username)
	lib.ShowFlex("header-user-info")

	resetLobby()
	lib.Show("mode-selection")
	lib.Hide("friend-mode-panel")
	lib.Hide("matchmaking-panel")
	lib.ShowScreen("lobby")
}

// handleGameCreated processes game created confirmation
func handleGameCreated(data interface{}) {
	var created lib.GameCreatedData
	if err := remarshal(data, &created); err != nil {
		lib.Console("handleGameCreated: remarshal failed: " + err.Error())
		return
	}

	lib.Get().SetGameCode(created.Code)
	showWaitingArea()
}

// handleGameStart processes game start message
func handleGameStart(data interface{}) {
	var start lib.GameStartData
	if err := remarshal(data, &start); err != nil {
		lib.Console("handleGameStart: remarshal failed: " + err.Error())
		return
	}

	state := lib.Get()
	state.SetGameCode(start.Code)
	state.SetCurrentTurn(start.CurrentTurn)
	state.SetPlayers(start.Players)
	state.SetReplayRequested(false)
	state.SetOpponentRequestedReplay(false)
	state.SetTimeRemaining(start.TimeRemaining)
	state.SetGameFinished(false)

	state.ResetBoard()
	state.ClearHover()
	state.FindPlayerIndex()

	updatePlayers()
	hideGameCode()
	hideReplayArea()
	showGameActions()
	lib.ShowScreen("game")
	lib.Draw()
	updateGameStatus()
	lib.Start()
}

// handleGameState processes full game state (reconnection)
func handleGameState(data interface{}) {
	var gameState lib.GameStateData
	if err := remarshal(data, &gameState); err != nil {
		lib.Console("handleGameState: remarshal failed: " + err.Error())
		return
	}

	state := lib.Get()
	state.SetGameCode(gameState.Code)
	state.SetCurrentTurn(gameState.CurrentTurn)
	state.SetBoard(gameState.Board)
	state.SetPlayers(gameState.Players)
	state.SetTimeRemaining(gameState.TimeRemaining)

	state.FindPlayerIndex()

	// Restore replay state if game is finished
	if gameState.Status == 2 {
		playerIdx := state.GetPlayerIdx()
		opponentIdx := 1 - playerIdx
		if playerIdx >= 0 && playerIdx < 2 {
			state.SetReplayRequested(gameState.ReplayRequests[playerIdx])
			state.SetOpponentRequestedReplay(gameState.ReplayRequests[opponentIdx])
		}
	}

	// Restore last move highlight
	if gameState.LastMove != nil {
		state.SetLastMove(gameState.LastMove.Col, gameState.LastMove.Row)
	}

	updatePlayers()

	switch gameState.Status {
	case 1: // Playing
		state.SetGameFinished(false)
		hideGameCode()
		hideWaitingActions()
		hideReplayArea()
		showGameActions()
		lib.ShowScreen("game")
		lib.Draw()
		updateGameStatus()
		lib.Start()

	case 0: // Waiting
		state.SetGameFinished(false)
		hideReplayArea()
		lib.ShowScreen("lobby")
		showWaitingActions()
		lib.Stop()

	case 2: // Finished
		state.SetGameFinished(true)
		hideGameCode()
		hideWaitingActions()
		showGameActions()
		lib.ShowScreen("game")
		lib.Draw()
		showGameOver(gameState.Result)
		lib.Stop()
	}
}

// handleMove processes opponent's move
func handleMove(data interface{}) {
	var move lib.MoveData
	if err := remarshal(data, &move); err != nil {
		lib.Console("handleMove: remarshal failed: " + err.Error())
		return
	}

	state := lib.Get()
	state.SetBoard(move.Board)
	state.SetCurrentTurn(move.NextTurn)
	state.SetTimeRemaining(move.TimeRemaining)
	state.SetLastMove(move.Column, move.Row)

	playedBy := 1 - move.NextTurn
	lib.AnimateDrop(move.Column, move.Row, playedBy)
	updateGameStatus()
}

// handleGameOver processes game over message
func handleGameOver(data interface{}) {
	var gameOver lib.GameOverData
	if err := remarshal(data, &gameOver); err != nil {
		lib.Console("handleGameOver: remarshal failed: " + err.Error())
		return
	}

	state := lib.Get()
	state.SetBoard(gameOver.Board)
	state.SetGameFinished(true)
	lib.Draw()
	showGameOver(gameOver.Result)
	lib.Stop()
}

// handleReplayRequest processes replay request from opponent
func handleReplayRequest(data interface{}) {
	var req lib.ReplayRequestData
	if err := remarshal(data, &req); err != nil {
		lib.Console("handleReplayRequest: remarshal failed: " + err.Error())
		return
	}

	state := lib.Get()
	if req.PlayerIdx != state.GetPlayerIdx() {
		state.SetOpponentRequestedReplay(true)
		updateReplayButton()
	}
}

// handleMatchmakingSearching confirms matchmaking search started
func handleMatchmakingSearching(data interface{}) {
	lib.Console("handleMatchmakingSearching: processing")
}

// handleQueueUpdate processes matchmaking queue updates
func handleQueueUpdate(data interface{}) {
	var queueData struct {
		PlayersInQueue int `json:"players_in_queue"`
	}

	if err := remarshal(data, &queueData); err != nil {
		return
	}

	playerCount := queueData.PlayersInQueue
	countText := formatPlayerCount(playerCount)

	lib.SetText("player-count", countText)

	// Update matchmaking status if searching
	if lib.GetElement("matchmaking-searching").Get("style").Get("display").String() == "none" {
		statusText := getMatchmakingStatus(playerCount)
		lib.SetText("matchmaking-status", statusText)
	}
}

// handleError processes error messages
func handleError(data interface{}) {
	var errData lib.ErrorData
	if err := remarshal(data, &errData); err != nil {
		return
	}

	lib.ShowMessage("lobby-message", errData.Message, "error")

	// Auto-clear error message after delay
	time.AfterFunc(errorMessageDisplayTime, func() {
		clearMessage("lobby-message")
	})
}

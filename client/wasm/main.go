// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 08.12.2025
//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"syscall/js"
	"time"

	"github.com/marvinEgger/GOnnect4/client/wasm/lib"
)

func main() {
	lib.Console("GOnnect4 WASM client starting...")

	// Initialize
	lib.Initialize()
	setupEventListeners()
	setupGlobalFunctions()

	// Auto-reconnect if we have saved credentials
	savedPlayerID := lib.GetLocalStorage("playerID")
	savedUsername := lib.GetLocalStorage("username")

	lib.Console("Saved credentials: playerID=" + savedPlayerID + ", username=" + savedUsername)

	if savedPlayerID != "" && savedUsername != "" {
		lib.Console("Auto-connecting...")
		autoConnect(savedUsername, savedPlayerID)
	} else {
		lib.Console("No saved credentials, showing login screen")
		lib.ShowScreen("login")
	}

	// Keep the program running
	select {}
}

// setupGlobalFunctions exposes Go functions to JavaScript
func setupGlobalFunctions() {
	// playColumn is called from board canvas click handler
	js.Global().Set("playColumn", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			col := args[0].Int()
			lib.SendMessage("play", map[string]interface{}{
				"column": col,
			})
		}
		return nil
	}))
}

// setupEventListeners attaches all event listeners
func setupEventListeners() {
	// Login
	attachEventListener("connect-btn", "click", handleConnect)
	attachKeyPressListener("username-input", handleConnect)

	// Lobby - Mode Selection
	attachEventListener("friend-mode-btn", "click", handleFriendMode)
	attachEventListener("matchmaking-btn", "click", handleMatchmakingMode)
	attachEventListener("back-to-modes-btn", "click", handleBackToModes)
	attachEventListener("back-to-modes-matchmaking-btn", "click", handleBackToModes)
	attachEventListener("logout-btn-header", "click", handleLogout)

	// Friend Mode
	attachEventListener("create-game-btn", "click", handleCreateGame)
	attachEventListener("join-game-btn", "click", handleJoinGame)
	attachKeyPressListener("join-code-input", handleJoinGame)
	attachEventListener("copy-code-btn", "click", handleCopyCode)

	// Matchmaking Mode
	attachEventListener("cancel-matchmaking-btn", "click", handleCancelMatchmaking)

	// Game
	attachEventListener("copy-code-game-btn", "click", handleCopyGameCode)
	attachEventListener("replay-btn", "click", handleReplay)
	attachEventListener("forfeit-btn", "click", handleForfeit)
	attachEventListener("back-to-lobby-btn", "click", handleBackToLobby)
	attachEventListener("cancel-game-btn", "click", handleCancelGame)

	// Board
	canvas := lib.GetElement("game-board")
	if !canvas.IsNull() {
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
}

// attachEventListener attaches a simple click event listener
func attachEventListener(id, event string, handler func(js.Value, []js.Value) interface{}) {
	el := lib.GetElement(id)
	if !el.IsNull() {
		el.Call("addEventListener", event, js.FuncOf(handler))
	}
}

// attachKeyPressListener attaches an Enter key listener
func attachKeyPressListener(id string, handler func(js.Value, []js.Value) interface{}) {
	el := lib.GetElement(id)
	if !el.IsNull() {
		el.Call("addEventListener", "keypress", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			event := args[0]
			if event.Get("key").String() == "Enter" {
				handler(this, args)
			}
			return nil
		}))
	}
}

// Event Handlers

func handleConnect(this js.Value, args []js.Value) interface{} {
	username := lib.GetValue("username-input")
	if username == "" {
		lib.ShowMessage("login-message", "Please enter a username", "error")
		return nil
	}

	// Try to connect to server
	lib.Connect(username, "", handleMessage)

	return nil
}

func handleCreateGame(this js.Value, args []js.Value) interface{} {
	lib.SendMessage("create_game", map[string]interface{}{})
	showWaitingArea()
	return nil
}

func handleJoinGame(this js.Value, args []js.Value) interface{} {
	code := lib.GetValue("join-code-input")
	if code == "" {
		lib.ShowMessage("lobby-message", "Please enter a game code", "error")
		return nil
	}

	// Join the game with the code
	lib.SendMessage("join_game", map[string]interface{}{
		"code": code,
	})
	return nil
}

func handleCopyCode(this js.Value, args []js.Value) any {
	code := lib.Get().GetGameCode()

	js.Global().
		Get("navigator").
		Get("clipboard").
		Call("writeText", code)

	lib.ShowMessage("lobby-message", "Code copied!", "success")

	// Hide message after 3 seconds
	time.AfterFunc(3*time.Second, func() {
		lib.SetText("lobby-message", "")
		lib.GetElement("lobby-message").Set("className", "message")
	})

	return js.Undefined()
}

func handleCopyGameCode(this js.Value, args []js.Value) any {
	code := lib.Get().GetGameCode()
	js.Global().Get("navigator").Get("clipboard").Call("writeText", code)

	lib.SetText("copy-code-game-btn", "Copied!")
	time.AfterFunc(2*time.Second, func() {
		lib.SetText("copy-code-game-btn", "Copy")
	})

	return js.Undefined()
}

func handleReplay(this js.Value, args []js.Value) interface{} {
	state := lib.Get()
	state.SetReplayRequested(true)
	lib.SendMessage("replay", map[string]interface{}{})
	updateReplayButton()
	return nil
}

func handleForfeit(this js.Value, args []js.Value) interface{} {
	if lib.Confirm("Are you sure you want to forfeit? Your opponent will win.") {
		lib.SendMessage("forfeit", map[string]interface{}{})
	}
	return nil
}

func handleBackToLobby(this js.Value, args []js.Value) interface{} {
	lib.SendMessage("leave_lobby", map[string]interface{}{})
	return nil
}

func handleCancelGame(this js.Value, args []js.Value) interface{} {
	lib.SendMessage("leave_lobby", map[string]interface{}{})
	return nil
}

func handleLogout(this js.Value, args []js.Value) interface{} {
	// Clear saved credentials
	lib.RemoveLocalStorage("playerID")
	lib.RemoveLocalStorage("username")

	// Close websocket
	lib.Close()

	// Stop timer
	lib.Stop()

	// Reset state
	state := lib.Get()
	state.SetPlayerID("")
	state.SetGameCode("")
	state.SetPlayerIdx(-1)

	// Clear input and go to login
	lib.SetValue("username-input", "")
	lib.Hide("header-user-info")
	lib.ShowScreen("login")

	return nil
}

// Mode Selection Handlers

func handleFriendMode(this js.Value, args []js.Value) interface{} {
	lib.Hide("mode-selection")
	lib.Hide("matchmaking-panel")
	lib.Show("friend-mode-panel")
	return nil
}

func handleMatchmakingMode(this js.Value, args []js.Value) interface{} {
	lib.Hide("mode-selection")
	lib.Hide("friend-mode-panel")
	lib.Show("matchmaking-panel")

	// Show searching immediately
	lib.Show("matchmaking-searching")

	// Random delay between 500ms and 3000ms to mix up users
	minDelay := 500
	maxDelay := 3000
	randomDelay := minDelay + int(js.Global().Get("Math").Call("random").Float()*float64(maxDelay-minDelay))

	// Send join_matchmaking after random delay
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		lib.SendMessage("join_matchmaking", map[string]interface{}{})
		return nil
	}), randomDelay)

	return nil
}

func handleBackToModes(this js.Value, args []js.Value) interface{} {
	// Hide all panels
	lib.Hide("friend-mode-panel")
	lib.Hide("matchmaking-panel")
	lib.Hide("waiting-area")
	lib.Hide("matchmaking-searching")

	// Cancel matchmaking if active
	lib.SendMessage("leave_matchmaking", map[string]interface{}{})

	// Clear lobby message
	lib.SetText("lobby-message", "")
	lib.GetElement("lobby-message").Set("className", "message")

	// Show mode selection
	lib.Show("mode-selection")

	// Reset lobby state
	resetLobby()

	return nil
}

// Matchmaking Handlers

func handleCancelMatchmaking(this js.Value, args []js.Value) interface{} {
	lib.SendMessage("leave_matchmaking", map[string]interface{}{})

	// Go back to mode selection
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
		case <-time.After(3 * time.Second):
			close(timedOut)
			lib.Console("Auto-reconnect timeout - clearing saved credentials")
			lib.RemoveLocalStorage("playerID")
			lib.RemoveLocalStorage("username")
			lib.Close()
			lib.ShowScreen("login")
		case <-done:
			// connecté/activité reçue avant timeout
		}
	}()

	// Custom message handler that cancels timeout (safe close)
	customHandler := func(msg lib.Message) {
		// if already timeout, just ignore
		select {
		case <-timedOut:
			return
		default:
		}

		// stop the timeout (once !)
		once.Do(func() { close(done) })
		handleMessage(msg)
	}

	lib.Connect(username, savedPlayerID, customHandler)
}

// Message Handlers

func handleMessage(msg lib.Message) {
	lib.Console("Received: " + msg.Type)

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

func handleWelcome(data interface{}) {
	lib.Console("handleWelcome: processing")
	var welcome lib.WelcomeData
	if err := remarshal(data, &welcome); err != nil {
		lib.Console("handleWelcome: remarshal failed: " + err.Error())
		return
	}

	state := lib.Get()
	state.SetPlayerID(welcome.PlayerID)

	lib.SetLocalStorage("playerID", welcome.PlayerID)
	lib.SetLocalStorage("username", welcome.Username)

	// Update header with username
	lib.SetText("header-username", welcome.Username)
	lib.ShowFlex("header-user-info")

	resetLobby()
	// Show mode selection
	lib.Show("mode-selection")
	lib.Hide("friend-mode-panel")
	lib.Hide("matchmaking-panel")
	lib.ShowScreen("lobby")
}

func handleGameCreated(data interface{}) {
	lib.Console("handleGameCreated: processing")
	var created lib.GameCreatedData
	if err := remarshal(data, &created); err != nil {
		lib.Console("handleGameCreated: remarshal failed: " + err.Error())
		return
	}

	lib.Get().SetGameCode(created.Code)
	showWaitingArea()
}

func handleGameStart(data interface{}) {
	lib.Console("handleGameStart: processing")
	var start lib.GameStartData
	if err := remarshal(data, &start); err != nil {
		lib.Console("handleGameStart: remarshal failed: " + err.Error())
		return
	}
	lib.Console("handleGameStart: success")

	state := lib.Get()
	state.SetGameCode(start.Code)
	state.SetCurrentTurn(start.CurrentTurn)
	state.SetPlayers(start.Players)
	state.SetReplayRequested(false)
	state.SetOpponentRequestedReplay(false)
	state.SetTimeRemaining(start.TimeRemaining)

	// Reset board and hover
	state.ResetBoard()
	state.ClearHover()

	// Find our player index
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

func handleGameState(data interface{}) {
	lib.Console("handleGameState: processing")
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

	// Find our player index
	state.FindPlayerIndex()

	// Restore replay state
	if gameState.Status == 2 { // Finished
		myIdx := state.GetPlayerIdx()
		opponentIdx := 1 - myIdx
		if myIdx >= 0 && myIdx < 2 {
			state.SetReplayRequested(gameState.ReplayRequests[myIdx])
			state.SetOpponentRequestedReplay(gameState.ReplayRequests[opponentIdx])
		}
	}

	// Restore last move
	if gameState.LastMove != nil {
		state.SetLastMove(gameState.LastMove.Col, gameState.LastMove.Row)
	}

	updatePlayers()

	if gameState.Status == 1 {
		// Playing
		hideGameCode()
		hideWaitingActions()

		hideReplayArea()
		showGameActions()
		lib.ShowScreen("game")
		lib.Draw()
		updateGameStatus()
		lib.Start()
	} else if gameState.Status == 0 {
		// Waiting
		hideReplayArea()
		lib.ShowScreen("lobby")
		showWaitingActions()
		lib.Stop()
	} else if gameState.Status == 2 {
		// Finished
		hideGameCode()
		hideWaitingActions()
		showGameActions()
		lib.ShowScreen("game")
		lib.Draw()
		showGameOver(gameState.Result)
		lib.Stop()
	}
}

func handleMove(data interface{}) {
	lib.Console("handleMove: processing")
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

func handleGameOver(data interface{}) {
	lib.Console("handleGameOver: processing")
	var gameOver lib.GameOverData
	if err := remarshal(data, &gameOver); err != nil {
		lib.Console("handleGameOver: remarshal failed: " + err.Error())
		return
	}

	lib.Get().SetBoard(gameOver.Board)
	lib.Draw()
	showGameOver(gameOver.Result)
	lib.Stop()
}

func handleReplayRequest(data interface{}) {
	lib.Console("handleReplayRequest: processing")
	var req lib.ReplayRequestData
	if err := remarshal(data, &req); err != nil {
		lib.Console("handleReplayRequest: remarshal failed: " + err.Error())
		return
	}

	// Only set opponent requested if it's not us
	state := lib.Get()
	if req.PlayerIdx != state.GetPlayerIdx() {
		state.SetOpponentRequestedReplay(true)
		updateReplayButton()
	}
}

func handleError(data interface{}) {
	var errData lib.ErrorData
	if err := remarshal(data, &errData); err != nil {
		return
	}

	lib.ShowMessage("lobby-message", errData.Message, "error")
}

func handleMatchmakingSearching(data interface{}) {
	lib.Console("handleMatchmakingSearching: processing")
	// Just confirm we're searching - UI already updated when button clicked
}

func handleQueueUpdate(data interface{}) {
	var queueData struct {
		PlayersInQueue int `json:"players_in_queue"`
	}
	if err := remarshal(data, &queueData); err != nil {
		return
	}

	// Update player count display
	playerCount := queueData.PlayersInQueue
	var countText string
	if playerCount == 0 {
		countText = "0 players online"
	} else if playerCount == 1 {
		countText = "1 player online"
	} else {
		countText = fmt.Sprintf("%d players online", playerCount)
	}

	lib.SetText("player-count", countText)

	// Update matchmaking status if in searching mode
	if lib.GetElement("matchmaking-searching").Get("style").Get("display").String() == "none" {
		statusText := "Looking for available players..."
		if playerCount >= 2 {
			statusText = "Match found! Starting game..."
		} else if playerCount == 1 {
			statusText = "Waiting for one more player..."
		}
		lib.SetText("matchmaking-status", statusText)
	}
}

// UI Helper Functions

func updatePlayers() {
	state := lib.Get()
	players := state.GetPlayers()
	currentTurn := state.GetCurrentTurn()
	playerIdx := state.GetPlayerIdx()

	for i := 0; i < 2; i++ {
		cardID := "player-" + string(rune('0'+i))

		// Update name
		name := players[i].Username
		if name == "" {
			name = "Waiting..."
		}
		nameEl := lib.GetElement(cardID)
		if !nameEl.IsNull() {
			nameDiv := nameEl.Call("querySelector", ".player-name")
			if !nameDiv.IsNull() {
				nameDiv.Set("textContent", name)
			}

			badgeDiv := nameEl.Call("querySelector", ".player-badge")
			if !badgeDiv.IsNull() {
				badge := "Opponent"
				if i == playerIdx {
					badge = "You"
				}
				badgeDiv.Set("textContent", badge)
			}
		}

		// Active state
		lib.ToggleClass(cardID, "active", currentTurn == i)
	}
}

func updateGameStatus() {
	state := lib.Get()

	if state.IsMyTurn() {
		lib.SetText("game-status", "Your turn - Click a column to play")
		lib.SetStyle("game-status", "color", "var(--success)")
	} else {
		lib.SetText("game-status", "Opponent's turn")
		lib.SetStyle("game-status", "color", "var(--text-secondary)")
	}
}

func showGameOver(result int) {
	state := lib.Get()
	playerIdx := state.GetPlayerIdx()

	var message, color string

	if result == 1 {
		// Player 0 win
		if playerIdx == 0 {
			message = "You won!"
			color = "var(--success)"
		} else {
			message = "You lost"
			color = "var(--danger)"
		}
	} else if result == 2 {
		// Player 1 win
		if playerIdx == 1 {
			message = "You won!"
			color = "var(--success)"
		} else {
			message = "You lost"
			color = "var(--danger)"
		}
	} else if result == 3 {
		// Draw
		message = "Draw"
		color = "var(--text-secondary)"
	}

	lib.SetText("game-status", message)
	lib.SetStyle("game-status", "color", color)

	hideGameActions()
	showReplayArea()
}

func showWaitingArea() {
	state := lib.Get()
	code := state.GetGameCode()

	lib.SetText("game-code-display", code)
	lib.Show("waiting-area")
	lib.Hide("create-game-btn")

	separator := js.Global().Get("document").Call("querySelector", ".separator")
	if !separator.IsNull() {
		separator.Get("style").Set("display", "none")
	}

	joinSection := js.Global().Get("document").Call("querySelector", ".lobby-section:last-of-type")
	if !joinSection.IsNull() {
		joinSection.Get("style").Set("display", "none")
	}
}

func resetLobby() {
	// Reset friend mode panel
	lib.Hide("waiting-area")
	lib.Show("create-game-btn")

	separator := js.Global().Get("document").Call("querySelector", ".separator")
	if !separator.IsNull() {
		separator.Get("style").Set("display", "block")
	}

	joinSection := js.Global().Get("document").Call("querySelector", ".lobby-section:last-of-type")
	if !joinSection.IsNull() {
		joinSection.Get("style").Set("display", "block")
	}

	lib.SetValue("join-code-input", "")
	lib.Get().SetGameCode("")

	// Reset matchmaking panel
	lib.Hide("matchmaking-searching")
}

func showGameCodeInGame() {
	state := lib.Get()
	code := state.GetGameCode()

	lib.ShowFlex("game-code-area")
	lib.SetText("game-status", "Waiting for opponent...")
	lib.SetStyle("game-status", "color", "var(--warning)")
	lib.SetText("game-code-info", code)
}

func hideGameCode() {
	lib.Hide("game-code-area")
}

func showReplayArea() {
	lib.Show("replay-area")
	updateReplayButton()
}

func hideReplayArea() {
	lib.Hide("replay-area")
	state := lib.Get()
	state.SetReplayRequested(false)
	state.SetOpponentRequestedReplay(false)
}

func showGameActions() {
	lib.ShowFlex("game-actions")
	lib.Hide("waiting-actions")
}

func hideGameActions() {
	lib.Hide("game-actions")
}

func showWaitingActions() {
	lib.ShowFlex("waiting-actions")
	lib.Hide("game-actions")
}

func hideWaitingActions() {
	lib.Hide("waiting-actions")
}

func updateReplayButton() {
	state := lib.Get()
	replayRequested := state.IsReplayRequested()
	opponentRequested := state.IsOpponentRequestedReplay()

	btn := lib.GetElement("replay-btn")
	if btn.IsNull() {
		return
	}

	if replayRequested && opponentRequested {
		btn.Set("textContent", "Restarting...")
		btn.Set("disabled", true)
	} else if replayRequested {
		btn.Set("textContent", "Waiting for opponent...")
		btn.Set("disabled", true)
	} else if opponentRequested {
		btn.Set("textContent", "Accept Replay")
		btn.Set("disabled", false)
		lib.AddClass("replay-btn", "btn-success")
		lib.RemoveClass("replay-btn", "btn-primary")
	} else {
		btn.Set("textContent", "Request Replay")
		btn.Set("disabled", false)
		lib.AddClass("replay-btn", "btn-primary")
		lib.RemoveClass("replay-btn", "btn-success")
	}
}

// remarshal converts interface{} to struct via JSON
func remarshal(in interface{}, out interface{}) error {
	bytes, err := json.Marshal(in)
	if err != nil {
		lib.Console("Error marshaling in remarshal: " + err.Error())
		return err
	}
	err = json.Unmarshal(bytes, out)
	if err != nil {
		lib.Console("Error unmarshaling in remarshal: " + err.Error())
	}
	return err
}

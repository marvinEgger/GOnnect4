// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 08.12.2025
//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"syscall/js"

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

			// For demo: simulate local play
			s := lib.Get()
			if s.IsMyTurn() {
				board := s.GetBoard()
				// Find lowest empty row in column
				for row := lib.Rows - 1; row >= 0; row-- {
					if board[row][col] == 0 {
						board[row][col] = s.GetPlayerIdx() + 1
						s.SetBoard(board)
						// Switch turn
						nextTurn := 1 - s.GetCurrentTurn()
						s.SetCurrentTurn(nextTurn)
						lib.Draw()
						updateGameStatus()
						break
					}
				}
			}
		}
		return nil
	}))
}

// setupEventListeners attaches all event listeners
func setupEventListeners() {
	// Login
	attachEventListener("connect-btn", "click", handleConnect)
	attachKeyPressListener("username-input", handleConnect)

	// Lobby
	attachEventListener("create-game-btn", "click", handleCreateGame)
	attachEventListener("join-game-btn", "click", handleJoinGame)
	attachKeyPressListener("join-code-input", handleJoinGame)
	attachEventListener("copy-code-btn", "click", handleCopyCode)
	attachEventListener("logout-btn", "click", handleLogout)

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

	// For demo/testing: simulate welcome after short delay if no server response
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// If still on login screen, simulate welcome
		if lib.GetElement("login-screen").Get("classList").Call("contains", "active").Bool() {
			lib.Console("No server response, simulating welcome for UI testing")
			s := lib.Get()
			s.SetPlayerID("demo-player-id")
			lib.SetLocalStorage("playerID", "demo-player-id")
			lib.SetLocalStorage("username", username)
			lib.SetText("welcome-message", "Welcome, "+username+" !")
			resetLobby()
			lib.ShowScreen("lobby")
		}
		return nil
	}), 500)

	return nil
}

func handleCreateGame(this js.Value, args []js.Value) interface{} {
	lib.SendMessage("create_game", map[string]interface{}{})

	// For demo/testing: simulate game creation if no server
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		s := lib.Get()
		if s.GetGameCode() == "" {
			lib.Console("No server response, simulating game creation for UI testing")
			s.SetGameCode("DEMO42")
			showWaitingArea()
		}
		return nil
	}), 300)

	return nil
}

func handleJoinGame(this js.Value, args []js.Value) interface{} {
	code := lib.GetValue("join-code-input")
	if code == "" {
		lib.ShowMessage("lobby-message", "Please enter a game code", "error")
		return nil
	}

	lib.SendMessage("join_game", map[string]interface{}{
		"code": code,
	})

	// For demo/testing: simulate game start if no server
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Simulate game start
		lib.Console("No server response, simulating game start for UI testing")
		s := lib.Get()
		s.SetGameCode(code)
		s.SetCurrentTurn(0)
		s.SetPlayerIdx(1) // Joiner is player 1
		s.SetPlayers([2]lib.Player{
			{ID: "opponent-id", Username: "Opponent"},
			{ID: s.GetPlayerID(), Username: lib.GetLocalStorage("username")},
		})
		s.ResetBoard()
		s.ClearHover()

		updatePlayers()
		lib.ShowScreen("game")
		lib.Draw()
		updateGameStatus()
		return nil
	}), 300)

	return nil
}

func handleCopyCode(this js.Value, args []js.Value) interface{} {
	s := lib.Get()
	code := s.GetGameCode()

	clipboard := js.Global().Get("navigator").Get("clipboard")
	clipboard.Call("writeText", code).Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		lib.ShowMessage("lobby-message", "Code copied!", "success")
		return nil
	}))

	return nil
}

func handleCopyGameCode(this js.Value, args []js.Value) interface{} {
	s := lib.Get()
	code := s.GetGameCode()

	clipboard := js.Global().Get("navigator").Get("clipboard")
	clipboard.Call("writeText", code).Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		lib.SetText("copy-code-game-btn", "Copied!")
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			lib.SetText("copy-code-game-btn", "Copy")
			return nil
		}), 2000)
		return nil
	}))

	return nil
}

func handleReplay(this js.Value, args []js.Value) interface{} {
	s := lib.Get()
	s.SetReplayRequested(true)
	lib.SendMessage("replay", map[string]interface{}{})
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

	// For demo: go back to lobby directly
	lib.Get().SetGameCode("")
	resetLobby()
	lib.ShowScreen("lobby")
	return nil
}

func handleCancelGame(this js.Value, args []js.Value) interface{} {
	lib.SendMessage("leave_lobby", map[string]interface{}{})

	// For demo: go back to lobby directly
	lib.Get().SetGameCode("")
	resetLobby()
	lib.ShowScreen("lobby")
	return nil
}

func handleLogout(this js.Value, args []js.Value) interface{} {
	// Clear saved credentials
	lib.RemoveLocalStorage("playerID")
	lib.RemoveLocalStorage("username")

	// Close websocket
	lib.Close()

	// Reset state
	s := lib.Get()
	s.SetPlayerID("")
	s.SetGameCode("")
	s.SetPlayerIdx(-1)

	// Clear input and go to login
	lib.SetValue("username-input", "")
	lib.ShowScreen("login")

	return nil
}

// autoConnect attempts to reconnect with saved credentials
func autoConnect(username, savedPlayerID string) {
	// Set a timeout
	timeoutID := js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		lib.Console("Auto-reconnect timeout - clearing saved credentials")
		lib.RemoveLocalStorage("playerID")
		lib.RemoveLocalStorage("username")
		lib.Close()
		lib.ShowScreen("login")
		return nil
	}), 3000)

	// Custom message handler that clears timeout
	customHandler := func(msg lib.Message) {
		js.Global().Call("clearTimeout", timeoutID)
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
	case "error":
		handleError(msg.Data)
	}
}

func handleWelcome(data interface{}) {
	var welcome lib.WelcomeData
	if err := remarshal(data, &welcome); err != nil {
		return
	}

	s := lib.Get()
	s.SetPlayerID(welcome.PlayerID)

	lib.SetLocalStorage("playerID", welcome.PlayerID)
	lib.SetLocalStorage("username", welcome.Username)

	lib.SetText("welcome-message", "Welcome, "+welcome.Username+" !")

	resetLobby()
	lib.ShowScreen("lobby")
}

func handleGameCreated(data interface{}) {
	var created lib.GameCreatedData
	if err := remarshal(data, &created); err != nil {
		return
	}

	lib.Get().SetGameCode(created.Code)
	showWaitingArea()
}

func handleGameStart(data interface{}) {
	var start lib.GameStartData
	if err := remarshal(data, &start); err != nil {
		return
	}

	s := lib.Get()
	s.SetGameCode(start.Code)
	s.SetCurrentTurn(start.CurrentTurn)
	s.SetPlayers(start.Players)
	s.SetReplayRequested(false)
	s.SetOpponentRequestedReplay(false)

	// Reset board and hover
	s.ResetBoard()
	s.ClearHover()

	// Find our player index
	s.FindPlayerIndex()

	updatePlayers()
	lib.ShowScreen("game")
	lib.Draw()
	updateGameStatus()
}

func handleGameState(data interface{}) {
	var gameState lib.GameStateData
	if err := remarshal(data, &gameState); err != nil {
		return
	}

	s := lib.Get()
	s.SetGameCode(gameState.Code)
	s.SetCurrentTurn(gameState.CurrentTurn)
	s.SetBoard(gameState.Board)
	s.SetPlayers(gameState.Players)

	// Find our player index
	s.FindPlayerIndex()

	updatePlayers()

	if gameState.Status == 1 { // Playing
		hideWaitingActions()
		lib.ShowScreen("game")
		lib.Draw()
		updateGameStatus()
	} else if gameState.Status == 0 { // Waiting
		showGameCodeInGame()
		showWaitingActions()
		lib.ShowScreen("game")
	}
}

func handleMove(data interface{}) {
	var move lib.MoveData
	if err := remarshal(data, &move); err != nil {
		return
	}

	s := lib.Get()
	s.SetBoard(move.Board)
	s.SetCurrentTurn(move.NextTurn)

	lib.Draw()
	updateGameStatus()
}

func handleGameOver(data interface{}) {
	var gameOver lib.GameOverData
	if err := remarshal(data, &gameOver); err != nil {
		return
	}

	lib.Get().SetBoard(gameOver.Board)
	lib.Draw()
	showGameOver(gameOver.Result)
}

func handleReplayRequest(data interface{}) {
	lib.Get().SetOpponentRequestedReplay(true)
}

func handleError(data interface{}) {
	var errData lib.ErrorData
	if err := remarshal(data, &errData); err != nil {
		return
	}

	lib.ShowMessage("lobby-message", errData.Message, "error")
}

// UI Helper Functions

func updatePlayers() {
	s := lib.Get()
	players := s.GetPlayers()
	currentTurn := s.GetCurrentTurn()
	playerIdx := s.GetPlayerIdx()

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
	s := lib.Get()

	if s.IsMyTurn() {
		lib.SetText("game-status", "Your turn - Click a column to play")
		lib.SetStyle("game-status", "color", "var(--success)")
	} else {
		lib.SetText("game-status", "Opponent's turn")
		lib.SetStyle("game-status", "color", "var(--text-secondary)")
	}
}

func showGameOver(result int) {
	s := lib.Get()
	playerIdx := s.GetPlayerIdx()

	var message, color string

	if result == 1 { // Player 0 win
		if playerIdx == 0 {
			message = "You won!"
			color = "var(--success)"
		} else {
			message = "You lost"
			color = "var(--danger)"
		}
	} else if result == 2 { // Player 1 win
		if playerIdx == 1 {
			message = "You won!"
			color = "var(--success)"
		} else {
			message = "You lost"
			color = "var(--danger)"
		}
	} else if result == 3 { // Draw
		message = "Draw"
		color = "var(--text-secondary)"
	}

	lib.SetText("game-status", message)
	lib.SetStyle("game-status", "color", color)
}

func showWaitingArea() {
	s := lib.Get()
	code := s.GetGameCode()

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

// TODO
func resetLobby() {

}

func showGameCodeInGame() {
	s := lib.Get()
	code := s.GetGameCode()

	lib.ShowFlex("game-code-area")
	lib.SetText("game-status", "Waiting for opponent...")
	lib.SetStyle("game-status", "color", "var(--warning)")
	lib.SetText("game-code-info", code)
}

func showWaitingActions() {
	lib.ShowFlex("waiting-actions")
	lib.Hide("game-actions")
}

func hideWaitingActions() {
	lib.Hide("waiting-actions")
}

// remarshal converts interface{} to struct via JSON
func remarshal(in interface{}, out interface{}) error {
	bytes, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, out)
}

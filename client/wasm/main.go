// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author: Astrit Aslani astrit.aslani@gmail.com
// Created: 08.12.2025
//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/marvinEgger/GOnnect4/client/wasm/lib"
)

// main entry point for the WASM client
func main() {
	lib.Console("GOnnect4 WASM client starting...")

	lib.Initialize()
	setupEventListeners()
	setupGlobalFunctions()

	attemptAutoConnect()

	// Keep the program running
	select {}
}

// setupGlobalFunctions exposes Go functions to JavaScript
func setupGlobalFunctions() {
	js.Global().Set("playColumn", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			column := args[0].Int()
			lib.SendMessage("play", map[string]interface{}{
				"column": column,
			})
		}
		return nil
	}))
}

// attemptAutoConnect tries to reconnect with saved credentials
func attemptAutoConnect() {
	savedPlayerID := lib.GetLocalStorage("playerID")
	savedUsername := lib.GetLocalStorage("username")

	if savedPlayerID != "" && savedUsername != "" {
		lib.Console("Auto-connecting...")
		autoConnect(savedUsername, savedPlayerID)
	} else {
		lib.Console("No saved credentials, showing login screen")
		lib.ShowScreen("login")
	}
}

// ------------------- //
// UI update functions //
// ------------------- //

// updatePlayers refreshes player information display
func updatePlayers() {
	state := lib.Get()
	players := state.GetPlayers()
	playerIdx := state.GetPlayerIdx()

	for i := 0; i < 2; i++ {
		cardID := "player-" + string(rune('0'+i))

		// Update player name
		name := players[i].Username
		if name == "" {
			name = "Waiting..."
		}

		nameElement := lib.GetElement(cardID)
		if !nameElement.IsNull() {
			nameDiv := nameElement.Call("querySelector", ".player-name")
			if !nameDiv.IsNull() {
				nameDiv.Set("textContent", name)
			}

			// Update badge (You/Opponent)
			badgeDiv := nameElement.Call("querySelector", ".player-badge")
			if !badgeDiv.IsNull() {
				badge := "Opponent"
				if i == playerIdx {
					badge = "You"
				}
				badgeDiv.Set("textContent", badge)
			}
		}

		// Update active state (shows which player is YOU, not the current turn)
		if playerIdx == i {
			lib.AddClass(cardID, "active")
		} else {
			lib.RemoveClass(cardID, "active")
		}
	}
}

// updateGameStatus updates the game status message
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

// showGameOver displays game over message
func showGameOver(result int) {
	state := lib.Get()
	playerIdx := state.GetPlayerIdx()

	var message, color string

	switch result {
	case 1:
		// Player 0 wins
		if playerIdx == 0 {
			message = "You won!"
			color = "var(--success)"
		} else {
			message = "You lost"
			color = "var(--danger)"
		}
	case 2:
		// Player 1 wins
		if playerIdx == 1 {
			message = "You won!"
			color = "var(--success)"
		} else {
			message = "You lost"
			color = "var(--danger)"
		}
	case 3:
		// Draw
		message = "Draw"
		color = "var(--text-secondary)"
	}

	lib.SetText("game-status", message)
	lib.SetStyle("game-status", "color", color)

	hideGameActions()
	showReplayArea()
}

// updateReplayButton updates replay button text and state
func updateReplayButton() {
	state := lib.Get()
	replayRequested := state.IsReplayRequested()
	opponentRequested := state.IsOpponentRequestedReplay()

	button := lib.GetElement("replay-btn")
	if button.IsNull() {
		return
	}

	switch {
	case replayRequested && opponentRequested:
		button.Set("textContent", "Restarting...")
		button.Set("disabled", true)

	case replayRequested:
		button.Set("textContent", "Waiting for opponent...")
		button.Set("disabled", true)

	case opponentRequested:
		button.Set("textContent", "Accept Replay")
		button.Set("disabled", false)
		lib.AddClass("replay-btn", "btn-success")
		lib.RemoveClass("replay-btn", "btn-primary")

	default:
		button.Set("textContent", "Request Replay")
		button.Set("disabled", false)
		lib.AddClass("replay-btn", "btn-primary")
		lib.RemoveClass("replay-btn", "btn-success")
	}
}

// -------------------------------- //
// Screen / UI management functions //
// -------------------------------- //

// showWaitingArea displays waiting for opponent screen
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

// resetLobby resets lobby to initial state
func resetLobby() {
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

	lib.Hide("matchmaking-searching")
}

// hideGameCode hides the game code display
func hideGameCode() {
	lib.Hide("game-code-area")
}

// showReplayArea shows replay request area
func showReplayArea() {
	lib.Show("replay-area")
	updateReplayButton()
}

// hideReplayArea hides replay request area
func hideReplayArea() {
	lib.Hide("replay-area")
	state := lib.Get()
	state.SetReplayRequested(false)
	state.SetOpponentRequestedReplay(false)
}

// showGameActions shows game action buttons
func showGameActions() {
	lib.ShowFlex("game-actions")
	lib.Hide("waiting-actions")
}

// hideGameActions hides game action buttons
func hideGameActions() {
	lib.Hide("game-actions")
}

// showWaitingActions shows waiting screen action buttons
func showWaitingActions() {
	lib.ShowFlex("waiting-actions")
	lib.Hide("game-actions")
}

// hideWaitingActions hides waiting screen action buttons
func hideWaitingActions() {
	lib.Hide("waiting-actions")
}

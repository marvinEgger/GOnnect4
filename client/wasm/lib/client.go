// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 08.12.2025
//go:build js && wasm

package lib

import (
	"encoding/json"
	"syscall/js"
)

// Message represents a WebSocket message
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WelcomeData contains welcome message data
type WelcomeData struct {
	PlayerID string `json:"player_id"`
	Username string `json:"username"`
}

// GameCreatedData contains game created data
type GameCreatedData struct {
	Code string `json:"code"`
}

// GameStartData contains game start data
type GameStartData struct {
	Code          string    `json:"code"`
	CurrentTurn   int       `json:"current_turn"`
	Players       [2]Player `json:"players"`
	TimeRemaining [2]int64  `json:"time_remaining"`
}

// GameStateData contains full game state
type GameStateData struct {
	Code          string    `json:"code"`
	Status        int       `json:"status"`
	CurrentTurn   int       `json:"current_turn"`
	Board         [6][7]int `json:"board"`
	Players       [2]Player `json:"players"`
	TimeRemaining [2]int64  `json:"time_remaining"`
}

// MoveData contains move information
type MoveData struct {
	PlayerIdx     int       `json:"player_idx"`
	Column        int       `json:"column"`
	Row           int       `json:"row"`
	Board         [6][7]int `json:"board"`
	NextTurn      int       `json:"next_turn"`
	TimeRemaining [2]int64  `json:"time_remaining"`
}

// GameOverData contains game over information
type GameOverData struct {
	Result int       `json:"result"`
	Board  [6][7]int `json:"board"`
}

// ErrorData contains error information
type ErrorData struct {
	Message string `json:"message"`
}

var (
	ws             js.Value
	messageHandler func(Message)
)

// Connect establishes WebSocket connection
func Connect(username, playerID string, onMessage func(Message)) {
	messageHandler = onMessage

	protocol := "ws:"
	if js.Global().Get("location").Get("protocol").String() == "https:" {
		protocol = "wss:"
	}

	host := js.Global().Get("location").Get("host").String()
	wsURL := protocol + "//" + host + "/ws"

	Console("Connecting to " + wsURL)

	ws = js.Global().Get("WebSocket").New(wsURL)

	// OnOpen handler
	ws.Call("addEventListener", "open", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		Console("Connected to server")

		// Send login message
		loginData := map[string]interface{}{
			"username": username,
		}
		if playerID != "" {
			loginData["player_id"] = playerID
		}

		SendMessage("login", loginData)
		return nil
	}))

	// OnMessage handler
	ws.Call("addEventListener", "message", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		data := event.Get("data").String()

		var msg Message
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			Console("Error parsing message: " + err.Error())
			return nil
		}

		if messageHandler != nil {
			messageHandler(msg)
		}

		return nil
	}))

	// OnError handler
	ws.Call("addEventListener", "error", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		Console("WebSocket error")
		ShowMessage("login-message", "Connection error", "error")
		return nil
	}))

	// OnClose handler
	ws.Call("addEventListener", "close", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		Console("Disconnected from server")
		return nil
	}))
}

// SendMessage sends a message to the server
func SendMessage(msgType string, data interface{}) {
	if ws.IsNull() || ws.IsUndefined() {
		Console("WebSocket not connected")
		return
	}

	readyState := ws.Get("readyState").Int()
	if readyState != 1 { // 1 = OPEN
		Console("WebSocket not ready")
		return
	}

	msg := Message{
		Type: msgType,
		Data: data,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		Console("Error marshaling message: " + err.Error())
		return
	}

	ws.Call("send", string(bytes))
}

// Close closes the WebSocket connection
func Close() {
	if !ws.IsNull() && !ws.IsUndefined() {
		ws.Call("close")
	}
}

// IsConnected checks if WebSocket is connected
func IsConnected() bool {
	if ws.IsNull() || ws.IsUndefined() {
		return false
	}
	return ws.Get("readyState").Int() == 1
}

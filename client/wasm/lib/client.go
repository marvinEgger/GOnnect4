package lib

import (
	js "syscall/js"
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

// TODO: Connect establishes WebSocket connection
func Connect(username, playerID string, onMessage func(Message)) {

}

// TODO: SendMessage sends a message to the server
func SendMessage(msgType string, data interface{}) {

}

// TODO: Close closes the WebSocket connection
func Close() {

}

// TODO: IsConnected checks if WebSocket is connected
func IsConnected() bool {
	return false
}

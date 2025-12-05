package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/marvinEgger/GOnnect4/client/wasm/lib"
)

func main() {

}

// TODO: setupGlobalFunctions exposes Go functions to JavaScript
func setupGlobalFunctions() {

}

// TODO: setupEventListeners attaches all event listeners
func setupEventListeners() {

}

// Event Handlers

// TODO
func handleConnect(this js.Value, args []js.Value) interface{} {
	return nil
}

// TODO
func handleCreateGame(this js.Value, args []js.Value) interface{} {
	return nil
}

// TODO
func handleJoinGame(this js.Value, args []js.Value) interface{} {
	return nil
}

// TODO
func handleReplay(this js.Value, args []js.Value) interface{} {
	return nil
}

// TODO
func handleForfeit(this js.Value, args []js.Value) interface{} {
	return nil
}

// TODO
func handleBackToLobby(this js.Value, args []js.Value) interface{} {
	return nil
}

// TODO
func handleCancelGame(this js.Value, args []js.Value) interface{} {
	return nil
}

// TODO
func handleLogout(this js.Value, args []js.Value) interface{} {
	return nil
}

// TODO: autoConnect attempts to reconnect with saved credentials
func autoConnect(username, savedPlayerID string) {

}

// Message Handlers

// TODO
func handleMessage(msg lib.Message) {

}

// TODO
func handleWelcome(data interface{}) {

}

// TODO
func handleGameCreated(data interface{}) {

}

// TODO
func handleGameStart(data interface{}) {

}

// TODO
func handleGameState(data interface{}) {

}

// TODO
func handleMove(data interface{}) {

}

// TODO
func handleGameOver(data interface{}) {

}

// TODO
func handleReplayRequest(data interface{}) {

}

// TODO
func handleError(data interface{}) {

}

// UI Helper Functions

// TODO
func updatePlayers() {

}

// TODO
func updateGameStatus() {

}

// TODO
func showGameOver(result int) {

}

// TODO
func showWaitingArea() {

}

// TODO
func resetLobby() {

}

// remarshal converts interface{} to struct via JSON
func remarshal(in interface{}, out interface{}) error {
	bytes, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, out)
}

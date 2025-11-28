package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/marvinEgger/GOnnect4/server/lib"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	defaultListenAddress = ":8080"
	webFolder            = "./client"
)

// Server manages all games and player connections
type Server struct {
	mu              sync.RWMutex
	gamesByCode     map[string]*lib.Game
	lobby           map[lib.PlayerID]*lib.Player
	bindingBySocket map[*websocket.Conn]*Binding
}

// Binding links a websocket connection to a player and game
type Binding struct {
	GameCode string
	PlayerID lib.PlayerID
}

// TODO: NewServer creates a new game server
func NewServer() *Server {
	return nil
}

// writeJSON sends a JSON message to a websocket
func writeJSON(ctx context.Context, conn *websocket.Conn, msg lib.Message) error {
	return wsjson.Write(ctx, conn, msg)
}

// TODO: sendError sends an error message to a client
func (s *Server) sendError(ctx context.Context, conn *websocket.Conn, message string) {

}

// TODO: handleLogin processes login/reconnection
func (s *Server) handleLogin(ctx context.Context, conn *websocket.Conn, binding *Binding, data lib.LoginData) {

}

// TODO: handleCreateGame creates a new game
func (s *Server) handleCreateGame(ctx context.Context, conn *websocket.Conn, binding *Binding) {

}

// TODO: handleJoinGame joins an existing game
func (s *Server) handleJoinGame(ctx context.Context, conn *websocket.Conn, binding *Binding, data lib.JoinGameData) {

}

// TODO: handlePlay processes a move
func (s *Server) handlePlay(ctx context.Context, conn *websocket.Conn, binding *Binding, data lib.PlayData) {

}

// TODO: handleReplay processes replay request
func (s *Server) handleReplay(ctx context.Context, conn *websocket.Conn, binding *Binding) {

}

// TODO: handleForfeit processes forfeit request
func (s *Server) handleForfeit(ctx context.Context, conn *websocket.Conn, binding *Binding) {

}

// TODO: handleLeaveLobby processes leave lobby request
func (s *Server) handleLeaveLobby(ctx context.Context, conn *websocket.Conn, binding *Binding) {

}

// TODO: sendGameState sends current game state to a player
func (s *Server) sendGameState(ctx context.Context, conn *websocket.Conn, game *lib.Game, playerID lib.PlayerID) {

}

// TODO: buildGameState constructs game state data
func (s *Server) buildGameState(game *lib.Game, playerID lib.PlayerID) lib.GameStateData {
	return lib.GameStateData{}
}

// TODO: getPlayerInfos gets public info for both players
func (s *Server) getPlayerInfos(game *lib.Game) [2]lib.PlayerInfo {
	return [2]lib.PlayerInfo{}
}

// TODO: getTimeRemaining gets remaining time for both players in milliseconds
func (s *Server) getTimeRemaining(game *lib.Game) [2]int64 {
	return [2]int64{}
}

// TODO: handleTimeout is called when a player's timer expires
func (s *Server) handleTimeout(gameCode string, loserIdx int) {

}

// TODO: broadcastToGame sends a message to all players in a game
func (s *Server) broadcastToGame(game *lib.Game, msg lib.Message) {

}

// TODO: cleanupStaleGames removes finished games and disconnected players
func (s *Server) cleanupStaleGames() {

}

// TODO: handleWebSocket handles websocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {

}

// TODO: handleMessage routes messages to appropriate handlers
func (s *Server) handleMessage(ctx context.Context, conn *websocket.Conn, binding *Binding, msg lib.Message) {

}

// mapToStruct converts interface{} to struct via JSON
func mapToStruct(in interface{}, out interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func main() {
	server := NewServer()

	http.HandleFunc("/ws", server.handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir(webFolder)))

	addr := defaultListenAddress
	fmt.Printf("Server starting on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

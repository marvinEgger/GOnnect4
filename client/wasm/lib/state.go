package lib

import "sync"

// Constants
const (
	Rows = 6
	Cols = 7
)

// Player represents player information
type Player struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// State holds all game state (singleton pattern)
type State struct {
	mu sync.RWMutex

	PlayerID                string
	GameCode                string
	PlayerIdx               int
	CurrentTurn             int
	Board                   [Rows][Cols]int
	HoverCol                int
	Players                 [2]Player
	ReplayRequested         bool
	OpponentRequestedReplay bool
}

var instance *State
var once sync.Once

// TODO: Get returns the singleton state instance
func Get() *State {
	return nil
}

// TODO: ResetBoard clears the board
func (s *State) ResetBoard() {
}

// GetHoverCol returns current hover column
func (s *State) GetHoverCol() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.HoverCol
}

// SetHoverCol updates hover column
func (s *State) SetHoverCol(col int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HoverCol = col
}

// ClearHover removes hover preview
func (s *State) ClearHover() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HoverCol = -1
}

// TODO: IsMyTurn checks if it's our turn
func (s *State) IsMyTurn() bool {
	return false
}

// TODO: GetBoard returns a copy of the board
func (s *State) GetBoard() [Rows][Cols]int {
	return [Rows][Cols]int{}
}

// TODO: SetBoard updates the entire board
func (s *State) SetBoard(board [Rows][Cols]int) {

}

// TODO: GetPlayerIdx returns player index
func (s *State) GetPlayerIdx() int {
	return -1
}

// TODO: SetPlayerIdx updates player index
func (s *State) SetPlayerIdx(idx int) {

}

// TODO: FindPlayerIndex finds our player index by ID
func (s *State) FindPlayerIndex() int {
	return -1
}

// TODO: SetPlayerID updates player ID
func (s *State) SetPlayerID(id string) {

}

// TODO: GetPlayerID returns player ID
func (s *State) GetPlayerID() string {
	return ""
}

// TODO: SetGameCode updates game code
func (s *State) SetGameCode(code string) {

}

// TODO: GetGameCode returns game code
func (s *State) GetGameCode() string {
	return ""
}

// SetCurrentTurn updates current turn
func (s *State) SetCurrentTurn(turn int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentTurn = turn
}

// GetCurrentTurn returns current turn
func (s *State) GetCurrentTurn() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CurrentTurn
}

// SetPlayers updates players array
func (s *State) SetPlayers(players [2]Player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Players = players
}

// TODO: GetPlayers returns players array
func (s *State) GetPlayers() [2]Player {
	return [2]Player{}
}

// TODO: SetReplayRequested updates replay request status
func (s *State) SetReplayRequested(requested bool) {

}

// TODO: IsReplayRequested returns replay request status
func (s *State) IsReplayRequested() bool {
	return false
}

// TODO: SetOpponentRequestedReplay updates opponent's replay request
func (s *State) SetOpponentRequestedReplay(requested bool) {

}

// TODO: IsOpponentRequestedReplay returns opponent's replay request status
func (s *State) IsOpponentRequestedReplay() bool {
	return false
}

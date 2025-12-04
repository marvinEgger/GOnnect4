package lib

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

const codeLength = 5

// Cell represents the state of a board cell
type Cell uint8

const (
	CellEmpty Cell = iota
	CellPlayer0
	CellPlayer1
)

// GameStatus represents the current state of a game
type GameStatus uint8

const (
	StatusWaiting GameStatus = iota
	StatusPlaying
	StatusFinished
)

// GameResult represents the outcome of a finished game
type GameResult uint8

const (
	ResultNone GameResult = iota
	ResultPlayer0Win
	ResultPlayer1Win
	ResultDraw
)

// Game represents a Connect 4 game session
type Game struct {
	mu     sync.RWMutex
	Code   string
	Board  *Board
	Status GameStatus
	Result GameResult

	Players      [2]*Player
	CurrentTurn  int
	MoveCount    int
	LastPlayedAt time.Time
	CreatedAt    time.Time

	ReplayRequests [2]bool

	// Timer management
	InitialClock  time.Duration // Store initial clock for resets
	TimeRemaining [2]time.Duration
	TurnStartedAt time.Time
	Timer         *time.Timer
	TimerCallback func(string, int) // Called when timer expires with (gameCode, loserIdx)
}

// NewGame creates a new game with a random code
func NewGame(initialClock time.Duration) *Game {
	return &Game{
		Code:          randomCode(codeLength),
		Board:         NewBoard(),
		Status:        StatusWaiting,
		CreatedAt:     time.Now(),
		InitialClock:  initialClock,
		TimeRemaining: [2]time.Duration{initialClock, initialClock},
	}}

// AddPlayer adds a player to the game
func (game *Game) AddPlayer(player *Player) bool {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.Status != StatusWaiting {
		return false
	}

	for i := range game.Players {
		if game.Players[i] == nil {
			game.Players[i] = player
			if i == 1 {
				game.start()
			}
			return true
		}
	}

	return false

}

// TODO: start begins the game when both players are ready
func (g *Game) start() {

}

// TODO: startTimer starts the timer for the current player
func (g *Game) startTimer() {

}

// TODO: stopTimer stops the timer and updates remaining time
func (g *Game) stopTimer() {

}

// TODO: Play attempts to play a move in the given column
func (g *Game) Play(playerIdx, col int) error {
	return nil
}

// TODO: RequestReplay marks a player's desire to replay
func (g *Game) RequestReplay(playerIdx int) bool {
	return false
}

// TODO: reset resets the game for a new round
func (g *Game) reset() {

}

// TODO: Cleanup stops all timers and releases resources
func (g *Game) Cleanup() {

}

// TODO: GetTimeRemaining returns remaining time for both players adjusted for current turn
func (g *Game) GetTimeRemaining() [2]time.Duration {
	return [2]time.Duration{}
}

// GetPlayerIndex returns the index of the given player
func (g *Game) GetPlayerIndex(id PlayerID) int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for i, p := range g.Players {
		if p != nil && p.ID == id {
			return i
		}
	}
	return -1
}

// HasPlayer checks if a player is in this game
func (g *Game) HasPlayer(id PlayerID) bool {
	return g.GetPlayerIndex(id) >= 0
}

// IsFull checks if the game has 2 players
func (g *Game) IsFull() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Players[0] != nil && g.Players[1] != nil
}

// randomCode generates a random alphanumeric code
func randomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate random code: " + err.Error())
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes)
}

// newToken generates a random hex token
func newToken(length int) string {
	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		panic("failed to generate random token: " + err.Error())
	}
	return strings.ToUpper(hex.EncodeToString(buffer))
}

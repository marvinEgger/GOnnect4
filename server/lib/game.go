// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author: Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

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

// LastMove represents the coordinates of the last move
type LastMove struct {
	Col int `json:"col"`
	Row int `json:"row"`
}

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
	LastMove     *LastMove

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
	}
}

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

// start begins the game when both players are ready
func (g *Game) start() {
	// Randomize who starts
	g.CurrentTurn = randomFirstPlayer()

	g.Status = StatusPlaying
	g.TurnStartedAt = time.Now()
	g.LastPlayedAt = time.Now()
	g.startTimer()
}

// startTimer starts the timer for the current player
func (g *Game) startTimer() {
	if g.Timer != nil {
		g.Timer.Stop()
	}

	remaining := g.TimeRemaining[g.CurrentTurn]
	if remaining <= 0 {
		// Time already expired
		if g.TimerCallback != nil {
			g.TimerCallback(g.Code, g.CurrentTurn)
		}
		return
	}

	// Capture values to avoid race condition
	code := g.Code
	playerIndex := g.CurrentTurn
	g.Timer = time.AfterFunc(remaining, func() {
		if g.TimerCallback != nil {
			g.TimerCallback(code, playerIndex)
		}
	})
}

// stopTimer stops the timer and updates remaining time
func (g *Game) stopTimer() {
	if g.Timer != nil {
		g.Timer.Stop()
		elapsed := time.Since(g.TurnStartedAt)
		g.TimeRemaining[g.CurrentTurn] -= elapsed
		if g.TimeRemaining[g.CurrentTurn] < 0 {
			g.TimeRemaining[g.CurrentTurn] = 0
		}
	}
}

// Play attempts to play a move in the given column
func (g *Game) Play(playerIdx, col int) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Status != StatusPlaying {
		return ErrGameNotPlaying
	}

	if playerIdx != g.CurrentTurn {
		return ErrNotYourTurn
	}

	// Stop timer and update time
	g.stopTimer()

	player := Cell(int(CellPlayer0) + playerIdx)
	node, ok := g.Board.Play(col, player)
	if !ok {
		return ErrInvalidMove
	}

	g.MoveCount++
	g.LastPlayedAt = time.Now()
	g.LastMove = &LastMove{Col: node.Col, Row: node.Row}

	// Check for win
	if g.Board.CheckWin(node) {
		g.Status = StatusFinished
		g.Result = GameResult(int(ResultPlayer0Win) + playerIdx)
		if g.Timer != nil {
			g.Timer.Stop()
		}
		return nil
	}

	// Check for draw
	if g.Board.IsFull() {
		g.Status = StatusFinished
		g.Result = ResultDraw
		if g.Timer != nil {
			g.Timer.Stop()
		}
		return nil
	}

	// Switch turn
	g.CurrentTurn = 1 - g.CurrentTurn
	g.TurnStartedAt = time.Now()
	g.startTimer()

	return nil
}

// Forfeit handles a player forfeiting the game
func (g *Game) Forfeit(loserIdx int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Status != StatusPlaying {
		return
	}

	if g.Timer != nil {
		g.Timer.Stop()
	}

	opponentIdx := 1 - loserIdx
	g.Status = StatusFinished
	g.Result = GameResult(opponentIdx + 1)
}

// RequestReplay marks a player's desire to replay
func (g *Game) RequestReplay(playerIdx int) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Status != StatusFinished {
		return false
	}

	g.ReplayRequests[playerIdx] = true

	// Both players agreed
	if g.ReplayRequests[0] && g.ReplayRequests[1] {
		g.reset()
		g.swapBeginningPlayer()
		return true
	}

	return false
}

// swapBeginningPlayer changes the turn order
func (g *Game) swapBeginningPlayer() {
	g.Players[0], g.Players[1] = g.Players[1], g.Players[0]
}

// reset resets the game for a new round
func (g *Game) reset() {
	if g.Timer != nil {
		g.Timer.Stop()
	}

	g.Board.Reset()
	g.Status = StatusPlaying
	g.Result = ResultNone
	g.CurrentTurn = 0
	g.MoveCount = 0
	g.ReplayRequests = [2]bool{false, false}
	g.TurnStartedAt = time.Now()
	g.LastPlayedAt = time.Now()
	g.LastMove = nil

	// Reset timers to initial clock value
	g.TimeRemaining[0] = g.InitialClock
	g.TimeRemaining[1] = g.InitialClock

	g.startTimer()
}

// Cleanup stops all timers and releases resources
func (g *Game) Cleanup() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Timer != nil {
		g.Timer.Stop()
		g.Timer = nil
	}
	g.TimerCallback = nil
}

// GetTimeRemaining returns remaining time for both players adjusted for current turn
func (g *Game) GetTimeRemaining() [2]time.Duration {
	g.mu.RLock()
	defer g.mu.RUnlock()

	times := g.TimeRemaining
	if g.Status == StatusPlaying {
		// Adjust for current player's elapsed time
		elapsed := time.Since(g.TurnStartedAt)
		times[g.CurrentTurn] -= elapsed
		if times[g.CurrentTurn] < 0 {
			times[g.CurrentTurn] = 0
		}
	}
	return times
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

// GetPlayers returns the players in the game safely
func (g *Game) GetPlayers() [2]*Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Players
}

// GetStatus returns the current game status
func (g *Game) GetStatus() GameStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Status
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

// randomFirstPlayer returns 0 or 1 randomly
func randomFirstPlayer() int {
	var b [1]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("failed to generate random player: " + err.Error())
	}
	return int(b[0] % 2)
}

// newToken generates a random hex token
func newToken(length int) string {
	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		panic("failed to generate random token: " + err.Error())
	}
	return strings.ToUpper(hex.EncodeToString(buffer))
}

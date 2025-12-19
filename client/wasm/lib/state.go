// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 08.12.2025
//go:build js && wasm

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

// LastMove represents the last move played
type LastMove struct {
	Col int
	Row int
}

// State holds all game state (singleton pattern)
type State struct {
	mutex sync.RWMutex

	PlayerID                string
	GameCode                string
	PlayerIdx               int
	CurrentTurn             int
	Board                   [Rows][Cols]int
	HoverCol                int
	Players                 [2]Player
	ReplayRequested         bool
	OpponentRequestedReplay bool
	TimeRemaining           [2]int64 // milliseconds
	LastMove                *LastMove
}

var instance *State
var once sync.Once

// Get returns the singleton state instance
func Get() *State {
	once.Do(func() {
		instance = &State{
			PlayerIdx: -1,
			HoverCol:  -1,
		}
	})
	return instance
}

// FindPlayerIndex finds our player index by ID
func (state *State) FindPlayerIndex() int {
	state.mutex.Lock()
	defer state.mutex.Unlock()

	for i := 0; i < len(state.Players); i++ {
		if state.Players[i].ID == state.PlayerID {
			state.PlayerIdx = i
			return i
		}
	}
	return -1
}

// IsMyTurn checks if it's our turn
func (state *State) IsMyTurn() bool {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.CurrentTurn == state.PlayerIdx
}

// ResetBoard clears the board
func (state *State) ResetBoard() {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.Board = [Rows][Cols]int{}
	state.LastMove = nil
}

// ClearHover removes hover preview
func (state *State) ClearHover() {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.HoverCol = -1
}

// GetHoverCol returns current hover column
func (state *State) GetHoverCol() int {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.HoverCol
}

// SetHoverCol updates hover column
func (state *State) SetHoverCol(col int) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.HoverCol = col
}

// GetBoard returns a copy of the board
func (state *State) GetBoard() [Rows][Cols]int {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.Board
}

// SetBoard updates the entire board
func (state *State) SetBoard(board [Rows][Cols]int) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.Board = board
}

// GetPlayerIdx returns player index
func (state *State) GetPlayerIdx() int {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.PlayerIdx
}

// SetPlayerIdx updates player index
func (state *State) SetPlayerIdx(idx int) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.PlayerIdx = idx
}

// GetPlayerID returns player ID
func (state *State) GetPlayerID() string {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.PlayerID
}

// SetPlayerID updates player ID
func (state *State) SetPlayerID(id string) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.PlayerID = id
}

// GetGameCode returns game code
func (state *State) GetGameCode() string {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.GameCode
}

// SetGameCode updates game code
func (state *State) SetGameCode(code string) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.GameCode = code
}

// GetCurrentTurn returns current turn
func (state *State) GetCurrentTurn() int {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.CurrentTurn
}

// SetCurrentTurn updates current turn
func (state *State) SetCurrentTurn(turn int) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.CurrentTurn = turn
}

// GetPlayers returns players array
func (state *State) GetPlayers() [2]Player {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.Players
}

// SetPlayers updates players array
func (state *State) SetPlayers(players [2]Player) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.Players = players
}

// SetReplayRequested updates replay request status
func (state *State) SetReplayRequested(requested bool) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.ReplayRequested = requested
}

// IsReplayRequested returns replay request status
func (state *State) IsReplayRequested() bool {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.ReplayRequested
}

// SetOpponentRequestedReplay updates opponent's replay request
func (state *State) SetOpponentRequestedReplay(requested bool) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.OpponentRequestedReplay = requested
}

// IsOpponentRequestedReplay returns opponent's replay request status
func (state *State) IsOpponentRequestedReplay() bool {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.OpponentRequestedReplay
}

// SetTimeRemaining updates time remaining
func (state *State) SetTimeRemaining(times [2]int64) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.TimeRemaining = times
}

// GetTimeRemaining returns time remaining
func (state *State) GetTimeRemaining() [2]int64 {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.TimeRemaining
}

// SetLastMove updates the last move played
func (state *State) SetLastMove(col, row int) {
	state.mutex.Lock()
	defer state.mutex.Unlock()
	state.LastMove = &LastMove{Col: col, Row: row}
}

// GetLastMove returns the last move played
func (state *State) GetLastMove() *LastMove {
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	return state.LastMove
}

// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 24.12.2025

package lib

import "testing"

// TestPlay_ValidMove tests a valid move
func TestPlay_ValidMove(t *testing.T) {
	game := NewGame(0)
	p1 := NewPlayer("Alice", 0)
	p2 := NewPlayer("Bob", 0)
	game.AddPlayer(p1)
	game.AddPlayer(p2)

	initialTurn := game.CurrentTurn

	// First player plays column 0
	err := game.Play(initialTurn, 0)
	if err != nil {
		t.Errorf("Valid move should not error: %v", err)
	}

	// Check move was applied
	if game.MoveCount != 1 {
		t.Errorf("Expected move count 1, got %d", game.MoveCount)
	}

	// Check turn switched
	if game.CurrentTurn == initialTurn {
		t.Error("Turn should have switched after move")
	}
}

// TestPlay_NotYourTurn tests playing out of turn
func TestPlay_NotYourTurn(t *testing.T) {
	game := NewGame(0)
	p1 := NewPlayer("Alice", 0)
	p2 := NewPlayer("Bob", 0)
	game.AddPlayer(p1)
	game.AddPlayer(p2)

	wrongPlayer := 1 - game.CurrentTurn

	err := game.Play(wrongPlayer, 0)
	if err != ErrNotYourTurn {
		t.Errorf("Expected ErrNotYourTurn, got %v", err)
	}

	// No move should have been applied
	if game.MoveCount != 0 {
		t.Error("Move count should still be 0")
	}
}

// TestPlay_InvalidColumn tests playing in invalid column
func TestPlay_InvalidColumn(t *testing.T) {
	game := NewGame(0)
	p1 := NewPlayer("Alice", 0)
	p2 := NewPlayer("Bob", 0)
	game.AddPlayer(p1)
	game.AddPlayer(p2)

	// Try invalid column
	err := game.Play(game.CurrentTurn, 99)
	if err != ErrInvalidMove {
		t.Errorf("Expected ErrInvalidMove, got %v", err)
	}

	// No move should have been applied
	if game.MoveCount != 0 {
		t.Error("Move count should still be 0")
	}
}

// TestPlay_GameNotPlaying tests playing when game not in playing state
func TestPlay_GameNotPlaying(t *testing.T) {
	game := NewGame(0)
	p1 := NewPlayer("Alice", 0)
	game.AddPlayer(p1)

	// Game is waiting (only 1 player)
	err := game.Play(0, 0)
	if err != ErrGameNotPlaying {
		t.Errorf("Expected ErrGameNotPlaying, got %v", err)
	}
}

// TestPlay_FullColumn tests playing in full column
func TestPlay_FullColumn(t *testing.T) {
	game := NewGame(0)
	p1 := NewPlayer("Alice", 0)
	p2 := NewPlayer("Bob", 0)
	game.AddPlayer(p1)
	game.AddPlayer(p2)

	// Fill column 0 (6 rows)
	for i := 0; i < 6; i++ {
		currentPlayer := game.CurrentTurn
		err := game.Play(currentPlayer, 0)
		if err != nil {
			t.Fatalf("Move %d should succeed: %v", i, err)
		}
	}

	// Try to play in full column
	err := game.Play(game.CurrentTurn, 0)
	if err != ErrInvalidMove {
		t.Errorf("Expected ErrInvalidMove for full column, got %v", err)
	}
}

// TestPlay_WinDetection tests that win is detected
func TestPlay_WinDetection(t *testing.T) {
	game := NewGame(0)
	p1 := NewPlayer("Alice", 0)
	p2 := NewPlayer("Bob", 0)
	game.AddPlayer(p1)
	game.AddPlayer(p2)

	// Force player 0 to start
	game.CurrentTurn = 0

	// Create horizontal win for player 0
	// P0 plays: 0, 1, 2, 3
	// P1 plays: 0, 1, 2 (blocking attempts but in different rows)

	game.Play(0, 0) // P0 at (5,0)
	game.Play(1, 0) // P1 at (4,0)
	game.Play(0, 1) // P0 at (5,1)
	game.Play(1, 1) // P1 at (4,1)
	game.Play(0, 2) // P0 at (5,2)
	game.Play(1, 2) // P1 at (4,2)
	err := game.Play(0, 3) // P0 at (5,3) - WINS

	if err != nil {
		t.Errorf("Win move should not error: %v", err)
	}

	if game.Status != StatusFinished {
		t.Error("Game should be finished after win")
	}

	if game.Result != ResultPlayer0Win {
		t.Errorf("Expected Player0 win, got %v", game.Result)
	}
}

// TestPlay_TurnSwitching tests that turns alternate correctly
func TestPlay_TurnSwitching(t *testing.T) {
	game := NewGame(0)
	p1 := NewPlayer("Alice", 0)
	p2 := NewPlayer("Bob", 0)
	game.AddPlayer(p1)
	game.AddPlayer(p2)

	firstTurn := game.CurrentTurn

	// Play one move
	game.Play(firstTurn, 0)

	if game.CurrentTurn == firstTurn {
		t.Error("Turn should have switched")
	}

	secondTurn := game.CurrentTurn
	if secondTurn != 1-firstTurn {
		t.Error("Turn should be opposite of first turn")
	}

	// Play another move
	game.Play(secondTurn, 1)

	if game.CurrentTurn != firstTurn {
		t.Error("Turn should have switched back")
	}
}

// TestPlay_LastMoveTracking tests that last move is recorded
func TestPlay_LastMoveTracking(t *testing.T) {
	game := NewGame(0)
	p1 := NewPlayer("Alice", 0)
	p2 := NewPlayer("Bob", 0)
	game.AddPlayer(p1)
	game.AddPlayer(p2)

	// Play in column 3
	game.Play(game.CurrentTurn, 3)

	if game.LastMove == nil {
		t.Fatal("LastMove should be set")
	}

	if game.LastMove.Col != 3 {
		t.Errorf("Expected LastMove.Col = 3, got %d", game.LastMove.Col)
	}

	if game.LastMove.Row != 5 {
		t.Errorf("Expected LastMove.Row = 5 (bottom), got %d", game.LastMove.Row)
	}
}

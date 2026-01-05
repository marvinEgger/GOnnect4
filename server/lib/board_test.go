// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Marvin Egger marvin.egger@hotmail.com
// Created: 02.01.2026

package lib

import "testing"

// TestBuildGraph tests the BuildGraph function of the Board struct
func TestBuildGraph(t *testing.T) {

	board := NewBoard()

	// 2. Verify Dimensions
	if len(board.nodes) != Rows {
		t.Errorf("Expected %d rows, got %d", Rows, len(board.nodes))
	}
	if len(board.nodes[0]) != Cols {
		t.Errorf("Expected %d cols, got %d", Cols, len(board.nodes[0]))
	}

	// 3. Test a Middle Node (Should have all 8 neighbors)
	midRow, midCol := 2, 3
	midNode := board.GetNode(midRow, midCol)

	// Define the expected neighbors for the middle node
	expectations := map[Direction]*Node{
		DirUp:        board.GetNode(midRow-1, midCol),
		DirDown:      board.GetNode(midRow+1, midCol),
		DirLeft:      board.GetNode(midRow, midCol-1),
		DirRight:     board.GetNode(midRow, midCol+1),
		DirUpLeft:    board.GetNode(midRow-1, midCol-1),
		DirUpRight:   board.GetNode(midRow-1, midCol+1),
		DirDownLeft:  board.GetNode(midRow+1, midCol-1),
		DirDownRight: board.GetNode(midRow+1, midCol+1),
	}

	for dir, expectedNode := range expectations {
		actual := midNode.GetNeighbor(dir)

		if actual == nil {
			t.Errorf("Middle Node (%d,%d): Missing neighbor in direction %v", midRow, midCol, dir)
		}
		if actual != expectedNode {
			t.Errorf("Middle Node (%d,%d): Neighbor %v points to wrong node", midRow, midCol, dir)
		}
	}

	// Test a Corner Node, Should NOT have neighbors: Up, Left, UpLeft, UpRight, DownLeft
	cornerNode := board.GetNode(0, 0)

	// Check valid connections for Top-Left
	if cornerNode.GetNeighbor(DirRight) != board.GetNode(0, 1) {
		t.Error("Corner (0,0) missing Right neighbor")
	}
	if cornerNode.GetNeighbor(DirDown) != board.GetNode(1, 0) {
		t.Error("Corner (0,0) missing Down neighbor")
	}
	if cornerNode.GetNeighbor(DirDownRight) != board.GetNode(1, 1) {
		t.Error("Corner (0,0) missing DownRight neighbor")
	}

	// Check invalid connections (Should be nil)
	invalidDirs := []Direction{DirUp, DirLeft, DirUpLeft, DirUpRight, DirDownLeft}
	for _, dir := range invalidDirs {
		if cornerNode.GetNeighbor(dir) != nil {
			t.Errorf("Corner (0,0) should NOT have neighbor in direction %v", dir)
		}
	}
}

// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 24.12.2025

package lib

import "testing"

// createHorizontalChain creates nodes linked horizontally
func createHorizontalChain(length int, owner Cell) []*Node {
	nodes := make([]*Node, length)
	for i := 0; i < length; i++ {
		nodes[i] = NewNode(0, i)
		nodes[i].SetOwner(owner)
	}

	// Link horizontally
	for i := 0; i < length-1; i++ {
		nodes[i].SetNeighbor(DirRight, nodes[i+1])
	}

	return nodes
}

// createVerticalChain creates nodes linked vertically
func createVerticalChain(length int, owner Cell) []*Node {
	nodes := make([]*Node, length)
	for i := 0; i < length; i++ {
		nodes[i] = NewNode(i, 0)
		nodes[i].SetOwner(owner)
	}

	// Link vertically
	for i := 0; i < length-1; i++ {
		nodes[i].SetNeighbor(DirDown, nodes[i+1])
	}

	return nodes
}

// createDiagonalUpRightChain creates nodes in diagonal
func createDiagonalUpRightChain(length int, owner Cell) []*Node {
	nodes := make([]*Node, length)
	for i := 0; i < length; i++ {
		nodes[i] = NewNode(length-1-i, i)
		nodes[i].SetOwner(owner)
	}

	// Link diagonally up-right
	for i := 0; i < length-1; i++ {
		nodes[i].SetNeighbor(DirUpRight, nodes[i+1])
	}

	return nodes
}

// createDiagonalDownRightChain creates nodes in diagonal
func createDiagonalDownRightChain(length int, owner Cell) []*Node {
	nodes := make([]*Node, length)
	for i := 0; i < length; i++ {
		nodes[i] = NewNode(i, i)
		nodes[i].SetOwner(owner)
	}

	// Link diagonally down-right
	for i := 0; i < length-1; i++ {
		nodes[i].SetNeighbor(DirDownRight, nodes[i+1])
	}

	return nodes
}

// TestCheckWin_Horizontal tests horizontal wins
func TestCheckWin_Horizontal(t *testing.T) {
	nodes := createHorizontalChain(4, CellPlayer0)

	// All nodes should detect win
	for i, node := range nodes {
		if !node.CheckWin(4) {
			t.Errorf("Node %d should detect horizontal win", i)
		}
	}
}

// TestCheckWin_HorizontalNoWin tests horizontal with only 3 (not enough)
func TestCheckWin_HorizontalNoWin(t *testing.T) {
	nodes := createHorizontalChain(3, CellPlayer0)

	// None should detect win
	for i, node := range nodes {
		if node.CheckWin(4) {
			t.Errorf("Node %d should NOT win with only 3 in row", i)
		}
	}
}

// TestCheckWin_Vertical tests vertical wins
func TestCheckWin_Vertical(t *testing.T) {
	nodes := createVerticalChain(4, CellPlayer1)

	// All nodes should detect win
	for i, node := range nodes {
		if !node.CheckWin(4) {
			t.Errorf("Node %d should detect vertical win", i)
		}
	}
}

// TestCheckWin_VerticalNoWin tests vertical with only 3 (not enough)
func TestCheckWin_VerticalNoWin(t *testing.T) {
	nodes := createVerticalChain(3, CellPlayer1)

	// None should detect win
	for i, node := range nodes {
		if node.CheckWin(4) {
			t.Errorf("Node %d should NOT win with only 3 in column", i)
		}
	}
}

// TestCheckWin_DiagonalUpRight tests diagonal
func TestCheckWin_DiagonalUpRight(t *testing.T) {
	nodes := createDiagonalUpRightChain(4, CellPlayer0)

	// All nodes should detect win
	for i, node := range nodes {
		if !node.CheckWin(4) {
			t.Errorf("Node %d should detect diagonal up-right win", i)
		}
	}
}

// TestCheckWin_DiagonalDownRight tests diagonal
func TestCheckWin_DiagonalDownRight(t *testing.T) {
	nodes := createDiagonalDownRightChain(4, CellPlayer1)

	// All nodes should detect win
	for i, node := range nodes {
		if !node.CheckWin(4) {
			t.Errorf("Node %d should detect diagonal down-right win", i)
		}
	}
}

// TestCheckWin_EmptyNode tests that empty nodes cannot win
func TestCheckWin_EmptyNode(t *testing.T) {
	node := NewNode(0, 0)

	if node.CheckWin(4) {
		t.Error("Empty node should NOT win")
	}
}

// TestCheckWin_DifferentOwners tests broken sequence by different owner
func TestCheckWin_DifferentOwners(t *testing.T) {
	nodes := createHorizontalChain(4, CellPlayer0)

	// Break sequence by changing middle node owner
	nodes[2].SetOwner(CellPlayer1)

	// No node should win (sequence broken)
	for i, node := range nodes {
		if node.CheckWin(4) {
			t.Errorf("Node %d should NOT win with broken sequence", i)
		}
	}
}

// TestCheckWin_ExactlyFour tests that 5 in a row also wins
func TestCheckWin_ExactlyFour(t *testing.T) {
	nodes := createHorizontalChain(5, CellPlayer0)

	// All should win (have at least 4)
	for i := 0; i < 5; i++ {
		if !nodes[i].CheckWin(4) {
			t.Errorf("Node %d in sequence of 5 should win", i)
		}
	}
}

// TestSetNeighbor_Bidirectional tests automatic bidirectional linking
func TestSetNeighbor_Bidirectional(t *testing.T) {
	node1 := NewNode(0, 0)
	node2 := NewNode(0, 1)

	node1.SetNeighbor(DirRight, node2)

	// Check bidirectional link
	if node1.GetNeighbor(DirRight) != node2 {
		t.Error("node1 should have node2 as right neighbor")
	}
	if node2.GetNeighbor(DirLeft) != node1 {
		t.Error("node2 should have node1 as left neighbor (automatic)")
	}
}

// TestIsEmpty tests empty node detection
func TestIsEmpty(t *testing.T) {
	node := NewNode(0, 0)

	if !node.IsEmpty() {
		t.Error("New node should be empty")
	}

	node.SetOwner(CellPlayer0)
	if node.IsEmpty() {
		t.Error("Node with owner should not be empty")
	}
}

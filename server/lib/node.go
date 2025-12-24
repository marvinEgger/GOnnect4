// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author: Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package lib

// Node represents a cell in the board as a graph node
type Node struct {
	Row       int
	Col       int
	Owner     Cell
	Neighbors [dirCount]*Node
}

// NewNode creates a new empty node at given position
func NewNode(row, col int) *Node {
	return &Node{
		Row:   row,
		Col:   col,
		Owner: CellEmpty,
	}
}

// GetNeighbor returns the neighbor in the given direction, or nil if none
func (n *Node) GetNeighbor(dir Direction) *Node {
	return n.Neighbors[dir]
}

// SetNeighbor sets the neighbor in the given direction
func (n *Node) SetNeighbor(dir Direction, neighbor *Node) {
	n.Neighbors[dir] = neighbor
	if neighbor != nil {
		neighbor.Neighbors[dir.Opposite()] = n
	}
}

// IsEmpty checks if the node has no owner
func (n *Node) IsEmpty() bool {
	return n.Owner == CellEmpty
}

// SetOwner sets the owner of this node
func (n *Node) SetOwner(player Cell) {
	n.Owner = player
}

// countSequence counts consecutive nodes with same owner in given direction
func (n *Node) countSequence(dir Direction) int {
	if n.IsEmpty() {
		return 0
	}

	count := 0
	current := n.GetNeighbor(dir)

	for current != nil && current.Owner == n.Owner {
		count++
		current = current.GetNeighbor(dir)
	}

	return count
}

// CheckWin checks if placing a token at this node creates a winning sequence
func (n *Node) CheckWin(winLength int) bool {

	if n.IsEmpty() {
		return false
	}

	// Check horizontal (left + right)
	if 1+n.countSequence(DirLeft)+n.countSequence(DirRight) >= winLength {
		return true
	}

	// Check vertical (up + down)
	if 1+n.countSequence(DirUp)+n.countSequence(DirDown) >= winLength {
		return true
	}

	// Check diagonal down-left to up-right
	if 1+n.countSequence(DirUpRight)+n.countSequence(DirDownLeft) >= winLength {
		return true
	}

	// Check diagonal up-left to down-right
	if 1+n.countSequence(DirUpLeft)+n.countSequence(DirDownRight) >= winLength {
		return true
	}

	return false
}

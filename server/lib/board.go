// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package lib

const (
	Rows      = 6
	Cols      = 7
	WinLength = 4
)

// Board represents the game board as a graph of connected nodes
type Board struct {
	nodes      [][]*Node
	colHeights [Cols]int
	rows       int
	cols       int
}

// NewBoard creates a new board and builds the node graph
func NewBoard() *Board {
	b := &Board{
		rows: Rows,
		cols: Cols,
	}
	b.buildGraph()
	return b
}

// buildGraph creates all nodes and establishes neighbor relationships
func (b *Board) buildGraph() {
	// Create all nodes
	b.nodes = make([][]*Node, b.rows)
	for row := 0; row < b.rows; row++ {
		b.nodes[row] = make([]*Node, b.cols)
		for col := 0; col < b.cols; col++ {
			b.nodes[row][col] = NewNode(row, col)
		}
	}

	// Establish neighbor relationships
	for row := 0; row < b.rows; row++ {
		for col := 0; col < b.cols; col++ {
			node := b.nodes[row][col]

			// Up
			if row > 0 {
				node.SetNeighbor(DirUp, b.nodes[row-1][col])
			}

			// Down
			if row < b.rows-1 {
				node.SetNeighbor(DirDown, b.nodes[row+1][col])
			}

			// Left
			if col > 0 {
				node.SetNeighbor(DirLeft, b.nodes[row][col-1])
			}

			// Right
			if col < b.cols-1 {
				node.SetNeighbor(DirRight, b.nodes[row][col+1])
			}

			// Diagonals
			if row > 0 && col > 0 {
				node.SetNeighbor(DirUpLeft, b.nodes[row-1][col-1])
			}

			if row > 0 && col < b.cols-1 {
				node.SetNeighbor(DirUpRight, b.nodes[row-1][col+1])
			}

			if row < b.rows-1 && col > 0 {
				node.SetNeighbor(DirDownLeft, b.nodes[row+1][col-1])
			}

			if row < b.rows-1 && col < b.cols-1 {
				node.SetNeighbor(DirDownRight, b.nodes[row+1][col+1])
			}
		}
	}

}

// canPlay checks if a column can accept a token
func (b *Board) canPlay(col int) bool {
	return col >= 0 && col < b.cols && b.colHeights[col] < b.rows
}

// Play drops a token in the given column for the given player
func (b *Board) Play(col int, player Cell) (*Node, bool) {
	if !b.canPlay(col) {
		return nil, false
	}

	row := b.rows - 1 - b.colHeights[col]
	node := b.nodes[row][col]
	node.SetOwner(player)
	b.colHeights[col]++

	return node, true
}

// CheckWin checks if the last played node creates a winning condition
func (b *Board) CheckWin(node *Node) bool {
	return node.CheckWin(WinLength)
}

// IsFull checks if the board is completely full
func (b *Board) IsFull() bool {
	for col := 0; col < b.cols; col++ {
		if b.colHeights[col] < b.rows {
			return false
		}
	}
	return true
}

// GetNode returns the node at given position
func (b *Board) GetNode(row, col int) *Node {
	if row < 0 || row >= b.rows || col < 0 || col >= b.cols {
		return nil
	}
	return b.nodes[row][col]
}

// GetLastPlayedNode returns the node at the top of a column
func (b *Board) GetLastPlayedNode(col int) *Node {
	if col < 0 || col >= b.cols || b.colHeights[col] == 0 {
		return nil
	}
	row := b.rows - b.colHeights[col]
	return b.nodes[row][col]
}

// Reset clears the board for a new game
func (b *Board) Reset() {
	for row := 0; row < b.rows; row++ {
		for col := 0; col < b.cols; col++ {
			b.nodes[row][col].SetOwner(CellEmpty)
		}
	}
	for col := 0; col < b.cols; col++ {
		b.colHeights[col] = 0
	}
}

// ToArray exports the board state as a 2D array
func (b *Board) ToArray() [Rows][Cols]Cell {
	var arr [Rows][Cols]Cell
	for row := 0; row < b.rows; row++ {
		for col := 0; col < b.cols; col++ {
			arr[row][col] = b.nodes[row][col].Owner
		}
	}
	return arr
}

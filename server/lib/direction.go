// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author: Astrit Aslani astrit.aslani@gmail.com
// Created: 05.12.2025

package lib

// Direction represents the 8 possible neighbor directions in the board graph
type Direction uint8

const (
	DirUp Direction = iota
	DirUpRight
	DirRight
	DirDownRight
	DirDown
	DirDownLeft
	DirLeft
	DirUpLeft
)

const dirCount = 8

// Opposite returns the opposite direction
func (d Direction) Opposite() Direction {
	return (d + 4) % dirCount
}

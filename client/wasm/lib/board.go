// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 08.12.2025
//go:build js && wasm

package lib

import "syscall/js"

// Constants
const (
	CellSize     = 80
	TokenRadius  = 28
	ShineRadius  = 8
	ShineOffset  = 8
	ShineAlpha   = 0.3
	PreviewAlpha = 0.4
)

// Token colors
const (
	ColorEmpty        = "#0f172a"
	ColorPlayer0      = "#dc2626"
	ColorPlayer1      = "#fbbf24"
	ColorPlayer0Alpha = "rgba(220, 38, 38, "
	ColorPlayer1Alpha = "rgba(251, 191, 36, "
	ColorBoardBg      = "#2962FF"
	ColorBoardBorder  = "#1d4ed8"
)

var (
	canvas         js.Value
	canevasContext js.Value
)

// Initialize sets up the canvas
func Initialize() {
	canvas = js.Global().Get("document").Call("getElementById", "game-board")
	if canvas.IsNull() {
		return
	}
	canevasContext = canvas.Call("getContext", "2d")
}

// Draw renders the entire board
func Draw() {
	if canevasContext.IsNull() {
		return
	}

	state := Get()
	board := state.GetBoard()
	width := canvas.Get("width").Int()
	height := canvas.Get("height").Int()

	// Clear
	canevasContext.Call("clearRect", 0, 0, width, height)

	// Background
	canevasContext.Set("fillStyle", ColorBoardBg)
	canevasContext.Call("fillRect", 0, 0, width, height)

	// Draw cells
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			x := col * CellSize
			y := row * CellSize

			// Cell border
			canevasContext.Set("strokeStyle", ColorBoardBorder)
			canevasContext.Set("lineWidth", 2)
			canevasContext.Call("strokeRect", x, y, CellSize, CellSize)

			// Token
			drawToken(x+CellSize/2, y+CellSize/2, board[row][col], 1.0)
		}
	}

	// Draw hover preview
	hoverCol := state.GetHoverCol()
	if hoverCol >= 0 && hoverCol < Cols && state.IsMyTurn() {
		drawHoverPreview(hoverCol, board)
	}
}

// drawHoverPreview draws ghost token preview
func drawHoverPreview(col int, board [Rows][Cols]int) {
	// Find lowest empty row
	row := -1
	for r := Rows - 1; r >= 0; r-- {
		if board[r][col] == 0 {
			row = r
			break
		}
	}

	if row >= 0 {
		x := col * CellSize
		y := row * CellSize
		playerToken := Get().GetPlayerIdx() + 1
		drawToken(x+CellSize/2, y+CellSize/2, playerToken, PreviewAlpha)
	}
}

// drawToken draws a single token
func drawToken(cx, cy, owner int, alpha float64) {
	canevasContext.Call("beginPath")
	canevasContext.Call("arc", cx, cy, TokenRadius, 0, 2*3.14159)

	switch owner {
	case 0:
		// Empty - dark hole
		canevasContext.Set("fillStyle", ColorEmpty)
	case 1:
		// Player 0 - Red
		if alpha < 1.0 {
			canevasContext.Set("fillStyle", ColorPlayer0Alpha+formatAlpha(alpha)+")")
		} else {
			canevasContext.Set("fillStyle", ColorPlayer0)
		}
	case 2:
		// Player 1 - Yellow
		if alpha < 1.0 {
			canevasContext.Set("fillStyle", ColorPlayer1Alpha+formatAlpha(alpha)+")")
		} else {
			canevasContext.Set("fillStyle", ColorPlayer1)
		}
	}

	canevasContext.Call("fill")

	// Shine effect for tokens
	if owner > 0 {
		canevasContext.Call("beginPath")
		canevasContext.Call("arc", cx-ShineOffset, cy-ShineOffset, ShineRadius, 0, 2*3.14159)
		shineAlpha := ShineAlpha * alpha
		canevasContext.Set("fillStyle", "rgba(255, 255, 255, "+formatAlpha(shineAlpha)+")")
		canevasContext.Call("fill")
	}
}

// formatAlpha formats alpha value for CSS
func formatAlpha(alpha float64) string {
	// Simple float to string conversion
	if alpha >= 1.0 {
		return "1"
	}
	if alpha <= 0.0 {
		return "0"
	}
	// Convert to string with 2 decimal places
	return js.Global().Get("Number").New(alpha).Call("toFixed", 2).String()
}

// HandleClick processes click on board
func HandleClick(event js.Value) {
	if !Get().IsMyTurn() {
		return
	}

	rect := canvas.Call("getBoundingClientRect")
	clientX := event.Get("clientX").Float()
	rectLeft := rect.Get("left").Float()
	rectWidth := rect.Get("width").Float()

	// Normalize x to canvas coordinates (0-560)
	x := (clientX - rectLeft) / rectWidth * 560.0
	col := int(x / CellSize)

	if col >= 0 && col < Cols {
		// Send play message via WebSocket
		// This will be called from main.go where we have access to WebSocket
		js.Global().Call("playColumn", col)
	}
}

// HandleHover updates hover column
func HandleHover(event js.Value) {
	if !Get().IsMyTurn() {
		return
	}

	rect := canvas.Call("getBoundingClientRect")
	clientX := event.Get("clientX").Float()
	rectLeft := rect.Get("left").Float()
	rectWidth := rect.Get("width").Float()

	// Normalize x to canvas coordinates (0-560)
	x := (clientX - rectLeft) / rectWidth * 560.0
	col := int(x / CellSize)

	state := Get()
	if col >= 0 && col < Cols && col != state.GetHoverCol() {
		state.SetHoverCol(col)
		Draw()
	}
}

// HandleLeave clears hover preview
func HandleLeave(event js.Value) {
	state := Get()
	if state.GetHoverCol() != -1 {
		state.ClearHover()
		Draw()
	}
}

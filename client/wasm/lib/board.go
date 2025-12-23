// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 08.12.2025
//go:build js && wasm

package lib

import "syscall/js"

// Board rendering constants
const (
	CellSize       = 80
	TokenRadius    = 28
	HighlightWidth = 8
	ShineRadius    = 8
	ShineOffset    = 8
	ShineAlpha     = 0.65
	PreviewAlpha   = 0.55
)

// Token colors
const (
	ColorEmpty        = "#0f172a"
	ColorPlayer0      = "#ce3262"
	ColorPlayer1      = "#fddd00"
	ColorPlayer0Alpha = "rgba(206, 48, 98, "
	ColorPlayer1Alpha = "rgba(253, 221, 0, "
	ColorBoardBg      = "#00add8"
	ColorBoardBorder  = "#5dc9e2"
	ColorHighlight    = "#5dc9e2"
)

// Animation constants
const (
	dropAnimationDuration = 550 // milliseconds
	dropStartY            = -TokenRadius * 2
)

var (
	// Main canvas (in DOM) - used for rendering
	canvas        js.Value
	canvasContext js.Value

	// Offscreen canvas - pre-rendered board overlay with holes
	// This optimization avoids redrawing the board frame every time and allows us to animate drops
	boardOverlayCanvas js.Value
	boardOverlayCtx    js.Value
)

// Initialize sets up the canvas and creates the board overlay
func Initialize() {
	canvas = js.Global().Get("document").Call("getElementById", "game-board")
	if canvas.IsNull() {
		return
	}
	canvasContext = canvas.Call("getContext", "2d")

	// Create offscreen canvas for board overlay
	boardOverlayCanvas = js.Global().Get("document").Call("createElement", "canvas")
	boardOverlayCanvas.Set("width", canvas.Get("width").Int())
	boardOverlayCanvas.Set("height", canvas.Get("height").Int())
	boardOverlayCtx = boardOverlayCanvas.Call("getContext", "2d")

	buildBoardOverlay()
}

// Draw renders the complete game board
func Draw() {
	if canvasContext.IsNull() {
		return
	}

	state := Get()
	board := state.GetBoard()
	lastMove := state.GetLastMove()
	canvasWidth := canvas.Get("width").Int()
	canvasHeight := canvas.Get("height").Int()

	canvasContext.Call("clearRect", 0, 0, canvasWidth, canvasHeight)

	// Draw all placed tokens
	drawPlacedTokens(board)

	// Draw hover preview if player's turn
	hoverColumn := state.GetHoverCol()
	if hoverColumn >= 0 && hoverColumn < Cols && state.IsMyTurn() {
		drawHoverPreview(hoverColumn, board)
	}

	// Draw board overlay (with holes)
	canvasContext.Call("drawImage", boardOverlayCanvas, 0, 0)

	// Draw highlight on last move
	if lastMove != nil {
		centerX := lastMove.Col*CellSize + CellSize/2
		centerY := lastMove.Row*CellSize + CellSize/2
		drawHighlight(centerX, centerY)
	}
}

// drawPlacedTokens renders all tokens currently on the board
func drawPlacedTokens(board [Rows][Cols]int) {
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			owner := board[row][col]
			if owner > 0 {
				centerX := col*CellSize + CellSize/2
				centerY := row*CellSize + CellSize/2
				drawToken(centerX, centerY, owner, 1.0)
			}
		}
	}
}

// drawHoverPreview draws ghost token preview at hover position
func drawHoverPreview(column int, board [Rows][Cols]int) {
	// Find lowest empty row
	targetRow := findLowestEmptyRow(column, board)

	if targetRow >= 0 {
		centerX := column*CellSize + CellSize/2
		centerY := targetRow*CellSize + CellSize/2
		playerToken := Get().GetPlayerIdx() + 1
		drawToken(centerX, centerY, playerToken, PreviewAlpha)
	}
}

// findLowestEmptyRow returns the lowest empty row in a column, or -1 if full
func findLowestEmptyRow(column int, board [Rows][Cols]int) int {
	for row := Rows - 1; row >= 0; row-- {
		if board[row][column] == 0 {
			return row
		}
	}
	return -1
}

// drawToken draws a single game token
func drawToken(centerX, centerY, owner int, alpha float64) {
	canvasContext.Call("beginPath")
	canvasContext.Call("arc", centerX, centerY, TokenRadius, 0, 2*3.14159)

	// Set token color based on owner
	switch owner {
	case 0:
		canvasContext.Set("fillStyle", ColorEmpty)
	case 1:
		if alpha < 1.0 {
			canvasContext.Set("fillStyle", ColorPlayer0Alpha+formatAlpha(alpha)+")")
		} else {
			canvasContext.Set("fillStyle", ColorPlayer0)
		}
	case 2:
		if alpha < 1.0 {
			canvasContext.Set("fillStyle", ColorPlayer1Alpha+formatAlpha(alpha)+")")
		} else {
			canvasContext.Set("fillStyle", ColorPlayer1)
		}
	}

	canvasContext.Call("fill")

	// Add shine effect for tokens
	if owner > 0 {
		drawShineEffect(centerX, centerY, alpha)
	}
}

// drawShineEffect adds a shine highlight to tokens
func drawShineEffect(centerX, centerY int, tokenAlpha float64) {
	canvasContext.Call("beginPath")
	canvasContext.Call("arc", centerX-ShineOffset, centerY-ShineOffset, ShineRadius, 0, 2*3.14159)
	shineAlpha := ShineAlpha * tokenAlpha
	canvasContext.Set("fillStyle", "rgba(255, 255, 255, "+formatAlpha(shineAlpha)+")")
	canvasContext.Call("fill")
}

// drawHighlight draws a ring around the last played token
func drawHighlight(centerX, centerY int) {
	canvasContext.Call("beginPath")
	canvasContext.Call("arc", centerX, centerY, TokenRadius, 0, 2*3.14159)
	canvasContext.Set("strokeStyle", ColorHighlight)
	canvasContext.Set("lineWidth", HighlightWidth)
	canvasContext.Call("stroke")
}

// drawFrameFalling renders a single animation frame during token drop
// The board state already contains the final token position, but we skip drawing it
// at its final location (excludeCol, excludeRow) to draw it at the animated position instead
func drawFrameFalling(excludeCol, excludeRow int, fallingX, fallingY float64, fallingOwner int) {
	state := Get()
	board := state.GetBoard()
	canvasWidth := canvas.Get("width").Int()
	canvasHeight := canvas.Get("height").Int()

	canvasContext.Call("clearRect", 0, 0, canvasWidth, canvasHeight)

	// Draw all placed tokens, skipping the one being animated
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			if col == excludeCol && row == excludeRow {
				continue
			}
			owner := board[row][col]
			if owner > 0 {
				centerX := col*CellSize + CellSize/2
				centerY := row*CellSize + CellSize/2
				drawToken(centerX, centerY, owner, 1.0)
			}
		}
	}

	// Draw falling token at its current animated position
	drawToken(int(fallingX), int(fallingY), fallingOwner, 1.0)

	// Draw board overlay (holes and grid) on top of everything
	canvasContext.Call("drawImage", boardOverlayCanvas, 0, 0)
}

// buildBoardOverlay creates the pre-rendered board frame with holes
func buildBoardOverlay() {
	overlayWidth := boardOverlayCanvas.Get("width").Int()
	overlayHeight := boardOverlayCanvas.Get("height").Int()

	// Fill board background
	boardOverlayCtx.Call("clearRect", 0, 0, overlayWidth, overlayHeight)
	boardOverlayCtx.Set("fillStyle", ColorBoardBg)
	boardOverlayCtx.Call("fillRect", 0, 0, overlayWidth, overlayHeight)

	// Punch out holes using destination-out compositing
	boardOverlayCtx.Call("save")
	boardOverlayCtx.Set("globalCompositeOperation", "destination-out")
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			centerX := col*CellSize + CellSize/2
			centerY := row*CellSize + CellSize/2
			boardOverlayCtx.Call("beginPath")
			boardOverlayCtx.Call("arc", centerX, centerY, TokenRadius, 0, 2*3.14159)
			boardOverlayCtx.Call("fill")
		}
	}
	boardOverlayCtx.Call("restore")

	// Draw grid lines
	drawGridLines()

	// Draw hole shadows for depth effect
	drawHoleShadows()
}

// drawGridLines draws the board grid
func drawGridLines() {
	boardOverlayCtx.Set("strokeStyle", ColorBoardBorder)
	boardOverlayCtx.Set("lineWidth", 2)
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			x := col * CellSize
			y := row * CellSize
			boardOverlayCtx.Call("strokeRect", x, y, CellSize, CellSize)
		}
	}
}

// drawHoleShadows adds subtle shadows around holes for depth
func drawHoleShadows() {
	boardOverlayCtx.Set("strokeStyle", "rgba(0,0,0,0.25)")
	boardOverlayCtx.Set("lineWidth", 4)
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			centerX := col*CellSize + CellSize/2
			centerY := row*CellSize + CellSize/2
			boardOverlayCtx.Call("beginPath")
			boardOverlayCtx.Call("arc", centerX, centerY, TokenRadius-2, 0, 2*3.14159)
			boardOverlayCtx.Call("stroke")
		}
	}
}

// AnimateDrop creates a drop animation for a token falling into position
// Uses requestAnimationFrame to create smooth 60fps animation with quadratic easing
// Animation flow :
//  1. Token starts above the board (dropStartY)
//  2. Falls to final position (row) over dropAnimationDuration ms
//  3. Uses quadratic easing (progress²) to simulate gravity acceleration
//  4. Calls Draw() when complete to render final state with highlight
func AnimateDrop(column, row, playerIdx int) {
	if canvasContext.IsNull() || canvas.IsNull() {
		Draw()
		return
	}

	// Step 1
	centerX := float64(column*CellSize + CellSize/2)
	endY := float64(row*CellSize + CellSize/2)
	startTime := js.Global().Get("performance").Call("now").Float()
	owner := playerIdx + 1

	var animate js.Func
	animate = js.FuncOf(func(this js.Value, args []js.Value) any {
		// Step 2
		currentTime := args[0].Float()
		progress := (currentTime - startTime) / dropAnimationDuration

		// Clamp progress to [0, 1]
		if progress < 0 {
			progress = 0
		}
		if progress > 1 {
			progress = 1
		}

		// Step 3 (Quadratic easing progress² gives gravity-like acceleration)
		eased := progress * progress
		currentY := dropStartY + (endY-dropStartY)*eased

		// Draw the falling token in canvas
		drawFrameFalling(column, row, centerX, currentY, owner)

		// Continue animation or finish
		if progress < 1 {
			js.Global().Call("requestAnimationFrame", animate)
		} else {
			// Step 4 final
			animate.Release()
			Draw()
		}
		return nil
	})

	js.Global().Call("requestAnimationFrame", animate)
}

// formatAlpha formats alpha value for CSS rgba
func formatAlpha(alpha float64) string {
	if alpha >= 1.0 {
		return "1"
	}
	if alpha <= 0.0 {
		return "0"
	}
	return js.Global().Get("Number").New(alpha).Call("toFixed", 2).String()
}

// HandleClick processes mouse clicks on the board
func HandleClick(event js.Value) {
	state := Get()

	// Ignore clicks when game is finished or not player's turn
	if state.GetGameFinished() || !state.IsMyTurn() {
		return
	}

	column := getColumnFromEvent(event)
	if column >= 0 && column < Cols {
		js.Global().Call("playColumn", column)
	}
}

// HandleHover updates hover column preview
func HandleHover(event js.Value) {
	state := Get()

	// Hide hover when game is finished or not player's turn
	if state.GetGameFinished() || !state.IsMyTurn() {
		return
	}

	column := getColumnFromEvent(event)

	if column >= 0 && column < Cols && column != state.GetHoverCol() {
		state.SetHoverCol(column)
		Draw()
	}
}

// HandleLeave clears hover preview when mouse leaves board
func HandleLeave(event js.Value) {
	state := Get()
	if state.GetHoverCol() != -1 {
		state.ClearHover()
		Draw()
	}
}

// getColumnFromEvent extracts the column number from a mouse event
func getColumnFromEvent(event js.Value) int {
	rect := canvas.Call("getBoundingClientRect")
	clientX := event.Get("clientX").Float()
	rectLeft := rect.Get("left").Float()
	rectWidth := rect.Get("width").Float()

	// Normalize x to canvas coordinates (0-560)
	x := (clientX - rectLeft) / rectWidth * 560.0
	return int(x / CellSize)
}

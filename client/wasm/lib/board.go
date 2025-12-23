// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 08.12.2025
//go:build js && wasm

package lib

import "syscall/js"

// Constants
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

var (
	canvas         js.Value
	canevasContext js.Value

	boardOverlayCanvas js.Value
	boardOverlayCtx    js.Value
)

// Initialize sets up the canvas
func Initialize() {
	canvas = js.Global().Get("document").Call("getElementById", "game-board")
	if canvas.IsNull() {
		return
	}
	canevasContext = canvas.Call("getContext", "2d")

	// Offscreen overlay (board + trous)
	boardOverlayCanvas = js.Global().Get("document").Call("createElement", "canvas")
	boardOverlayCanvas.Set("width", canvas.Get("width").Int())
	boardOverlayCanvas.Set("height", canvas.Get("height").Int())
	boardOverlayCtx = boardOverlayCanvas.Call("getContext", "2d")

	buildBoardOverlay()
}

// Draw renders the entire board
func Draw() {
	if canevasContext.IsNull() {
		return
	}

	state := Get()
	board := state.GetBoard()
	lastMove := state.GetLastMove()
	w := canvas.Get("width").Int()
	h := canvas.Get("height").Int()

	canevasContext.Call("clearRect", 0, 0, w, h)

	// Token played (without empty cell's)
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			owner := board[row][col]
			if owner > 0 {
				cx := col*CellSize + CellSize/2
				cy := row*CellSize + CellSize/2
				drawToken(cx, cy, owner, 1.0)
			}
		}
	}

	// Preview hover (draw behind the "board")
	hoverCol := state.GetHoverCol()
	if hoverCol >= 0 && hoverCol < Cols && state.IsMyTurn() {
		drawHoverPreview(hoverCol, board)
	}

	// Board par-dessus (avec vrais trous)
	canevasContext.Call("drawImage", boardOverlayCanvas, 0, 0)

	// Highlight (mets un rayon légèrement plus petit pour éviter de “déborder” sur la planche)
	if lastMove != nil {
		cx := lastMove.Col*CellSize + CellSize/2
		cy := lastMove.Row*CellSize + CellSize/2
		drawHighlight(cx, cy) // idéalement arc radius = TokenRadius - HighlightWidth/2
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

// drawHighlight draws a ring around the last played token
func drawHighlight(cx, cy int) {
	canevasContext.Call("beginPath")
	canevasContext.Call("arc", cx, cy, TokenRadius, 0, 2*3.14159)
	canevasContext.Set("strokeStyle", ColorHighlight)
	canevasContext.Set("lineWidth", HighlightWidth)
	canevasContext.Call("stroke")
}

func drawFrameFalling(excludeCol, excludeRow int, fallingX, fallingY float64, fallingOwner int) {
	state := Get()
	board := state.GetBoard()
	w := canvas.Get("width").Int()
	h := canvas.Get("height").Int()

	canevasContext.Call("clearRect", 0, 0, w, h)

	// Jetons posés (sauf la case finale du jeton animé)
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if c == excludeCol && r == excludeRow {
				continue
			}
			owner := board[r][c]
			if owner > 0 {
				drawToken(c*CellSize+CellSize/2, r*CellSize+CellSize/2, owner, 1.0)
			}
		}
	}

	// Jeton qui tombe
	drawToken(int(fallingX), int(fallingY), fallingOwner, 1.0)

	// Board overlay au-dessus
	canevasContext.Call("drawImage", boardOverlayCanvas, 0, 0)
}

func buildBoardOverlay() {
	w := boardOverlayCanvas.Get("width").Int()
	h := boardOverlayCanvas.Get("height").Int()

	// Fond board
	boardOverlayCtx.Call("clearRect", 0, 0, w, h)
	boardOverlayCtx.Set("fillStyle", ColorBoardBg)
	boardOverlayCtx.Call("fillRect", 0, 0, w, h)

	// Perce les trous (alpha = 0)
	boardOverlayCtx.Call("save")
	boardOverlayCtx.Set("globalCompositeOperation", "destination-out")
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			cx := col*CellSize + CellSize/2
			cy := row*CellSize + CellSize/2
			boardOverlayCtx.Call("beginPath")
			boardOverlayCtx.Call("arc", cx, cy, TokenRadius, 0, 2*3.14159)
			boardOverlayCtx.Call("fill")
		}
	}
	boardOverlayCtx.Call("restore")

	// Grille / bordures par-dessus
	boardOverlayCtx.Set("strokeStyle", ColorBoardBorder)
	boardOverlayCtx.Set("lineWidth", 2)
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			x := col * CellSize
			y := row * CellSize
			boardOverlayCtx.Call("strokeRect", x, y, CellSize, CellSize)
		}
	}

	// Petit “rebord” sombre autour des trous (optionnel mais joli)
	boardOverlayCtx.Set("strokeStyle", "rgba(0,0,0,0.25)")
	boardOverlayCtx.Set("lineWidth", 4)
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			cx := col*CellSize + CellSize/2
			cy := row*CellSize + CellSize/2
			boardOverlayCtx.Call("beginPath")
			boardOverlayCtx.Call("arc", cx, cy, TokenRadius-2, 0, 2*3.14159)
			boardOverlayCtx.Call("stroke")
		}
	}
}

// AnimateDrop animates a token falling into (col,row).
// playerIdx: 0 for player0, 1 for player1
func AnimateDrop(col, row, playerIdx int) {
	if canevasContext.IsNull() || canvas.IsNull() {
		Draw()
		return
	}

	// X center (column)
	x := float64(col*CellSize + CellSize/2)

	// Y: start above the board, end at cell center
	startY := float64(-TokenRadius * 2)
	endY := float64(row*CellSize + CellSize/2)

	// Duration in ms
	duration := 650.0
	startTime := js.Global().Get("performance").Call("now").Float()

	owner := playerIdx + 1 // board values are 1/2

	var animate js.Func
	animate = js.FuncOf(func(this js.Value, args []js.Value) any {
		now := args[0].Float()
		t := (now - startTime) / duration
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}

		// easing (gravity-ish)
		e := t * t
		y := startY + (endY-startY)*e

		// IMPORTANT: redraw a full frame (tokens + falling + overlay)
		drawFrameFalling(col, row, x, y, owner)

		if t < 1 {
			js.Global().Call("requestAnimationFrame", animate)
		} else {
			animate.Release()
			Draw() // final clean draw (with highlight etc.)
		}
		return nil
	})

	js.Global().Call("requestAnimationFrame", animate)
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

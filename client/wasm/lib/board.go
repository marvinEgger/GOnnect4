// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 08.12.2025
//go:build js && wasm
// +build js,wasm

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
	canvas js.Value
	ctx    js.Value
)

// Initialize sets up the canvas
func Initialize() {
	canvas = js.Global().Get("document").Call("getElementById", "game-board")
	if canvas.IsNull() {
		return
	}
	ctx = canvas.Call("getContext", "2d")
}

// Draw renders the entire board
func Draw() {
	if ctx.IsNull() {
		return
	}

	width := canvas.Get("width").Int()
	height := canvas.Get("height").Int()

	// Clear
	ctx.Call("clearRect", 0, 0, width, height)

	// Background
	ctx.Set("fillStyle", ColorBoardBg)
	ctx.Call("fillRect", 0, 0, width, height)

	// Draw cells
	for row := 0; row < Rows; row++ {
		for col := 0; col < Cols; col++ {
			x := col * CellSize
			y := row * CellSize

			// Cell border
			ctx.Set("strokeStyle", ColorBoardBorder)
			ctx.Set("lineWidth", 2)
			ctx.Call("strokeRect", x, y, CellSize, CellSize)
		}
	}
}

// TODO: drawHoverPreview draws ghost token preview
func drawHoverPreview(col int, board [Rows][Cols]int) {

}

// TODO: drawToken draws a single token
func drawToken(cx, cy, owner int, alpha float64) {

}

// TODO: formatAlpha formats alpha value for CSS
func formatAlpha(alpha float64) string {
	return ""
}

// TODO: HandleClick processes click on board
func HandleClick(event js.Value) {

}

// TODO: HandleHover updates hover column
func HandleHover(event js.Value) {

}

// TODO: HandleLeave clears hover preview
func HandleLeave(event js.Value) {

}

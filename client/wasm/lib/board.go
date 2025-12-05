package lib

import (
	js "syscall/js"
)

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

// TODO: Initialize sets up the canvas
func Initialize() {

}

// Draw renders the entire board
func Draw() {

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

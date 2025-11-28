package lib

import "errors"

var (
	ErrGameNotPlaying = errors.New("game is not in playing state")
	ErrNotYourTurn    = errors.New("not your turn")
	ErrInvalidMove    = errors.New("invalid move")
	ErrGameNotFound   = errors.New("game not found")
	ErrGameFull       = errors.New("game is full")
	ErrPlayerNotFound = errors.New("player not found")
)

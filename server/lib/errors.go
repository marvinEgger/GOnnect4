// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 05.12.2025

package lib

import "errors"

var (
	ErrGameNotPlaying      = errors.New("game is not in playing state")
	ErrNotYourTurn         = errors.New("not your turn")
	ErrInvalidMove         = errors.New("invalid move")
	ErrGameNotFound        = errors.New("game not found")
	ErrGameFull            = errors.New("game is full")
	ErrPlayerNotFound      = errors.New("player not found")
	ErrPlayerNotInGame     = errors.New("player not in game")
	ErrPlayerAlreadyInGame = errors.New("player already in game")
	ErrInvalidUsername     = errors.New("invalid username")
)

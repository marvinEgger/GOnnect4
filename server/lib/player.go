// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025
package lib

import (
	"strings"
	"sync"
	"time"
)

const tokenLength = 16

// PlayerID uniquely identifies a player session
type PlayerID string

// Player represents a connected player
type Player struct {
	sync.RWMutex
	ID        PlayerID
	Username  string
	Connected bool
	Remaining time.Duration
}

// NewPlayer creates a new player with a unique ID
func NewPlayer(username string, initialClock time.Duration) *Player {
	username = strings.TrimSpace(username)
	// check if username is empty
	if username == "" {
		return nil
	}

	// return the new player
	return &Player{

		ID:        PlayerID(newToken(tokenLength)),
		Username:  username,
		Connected: true,
		Remaining: initialClock,
	}
}

// SetConnected safely sets the connection status
func (p *Player) SetConnected(connected bool) {
	p.Lock()
	defer p.Unlock()
	p.Connected = connected
}

// IsConnected safely checks connection status
func (p *Player) IsConnected() bool {
	p.RLock()
	defer p.RUnlock()
	return p.Connected
}

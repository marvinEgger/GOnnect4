// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author: Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package lib

import (
	"strings"
	"sync"
	"time"
)

const tokenLength = 16

// Sender interface abstracts the network layer
type Sender interface {
	Send(Message)
}

// PlayerID uniquely identifies a player session
type PlayerID string

// Player represents a connected player
type Player struct {
	sync.RWMutex
	ID        PlayerID
	Username  string
	sender    Sender
	Remaining time.Duration
	LastSeen  time.Time
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
		Remaining: initialClock,
		LastSeen:  time.Now(),
	}
}

// SetSender sets the network sender for this player
func (p *Player) SetSender(s Sender) {
	p.Lock()
	defer p.Unlock()
	p.sender = s
	if s != nil {
		p.LastSeen = time.Now()
	}
}

// IsConnected checks if the player has an active sender
func (p *Player) IsConnected() bool {
	p.RLock()
	defer p.RUnlock()
	return p.sender != nil
}

// Send sends a message to the player if connected
func (p *Player) Send(msg Message) {
	p.RLock()
	defer p.RUnlock()
	if p.sender != nil {
		p.sender.Send(msg)
	}
}

// GetLastSeen returns the last seen timestamp
func (p *Player) GetLastSeen() time.Time {
	p.RLock()
	defer p.RUnlock()
	return p.LastSeen
}

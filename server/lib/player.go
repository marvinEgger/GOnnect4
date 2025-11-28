package lib

import (
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

// TODO: NewPlayer creates a new player with a unique ID
func NewPlayer(username string, initialClock time.Duration) *Player {
	return nil
}

// TODO: SetConnected safely sets the connection status
func (p *Player) SetConnected(connected bool) {

}

// TODO: IsConnected safely checks connection status
func (p *Player) IsConnected() bool {
	return false
}

// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 05.12.2025
package lib

import (
	"context"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// Client handles the websocket connection and implements lib.Sender
type Client struct {
	Conn     *websocket.Conn
	SendChan chan Message
	PlayerID PlayerID
	GameCode string
}

// NewClient creates a new client
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn:     conn,
		SendChan: make(chan Message, 256),
	}
}

// Send implements lib.Sender interface
func (c *Client) Send(msg Message) {
	select {
	case c.SendChan <- msg:
	default:
		// If channel is full, we might want to close connection or log warning
		// For now, we just drop to avoid blocking, but in production we should handle this better
		go c.Conn.Close(websocket.StatusPolicyViolation, "Connection too slow")
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case msg, ok := <-c.SendChan:
			if !ok {
				c.Conn.Close(websocket.StatusNormalClosure, "Channel closed")
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), writeWait)
			err := wsjson.Write(ctx, c.Conn, msg)
			cancel()
			if err != nil {
				return
			}

		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), writeWait)
			if err := c.Conn.Ping(ctx); err != nil {
				cancel()
				return
			}
			cancel()
		}
	}
}

// Close closes the send channel
func (c *Client) Close() {
	close(c.SendChan)
}

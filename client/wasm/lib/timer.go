// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 17.12.2025
//go:build js && wasm

package lib

import (
	"fmt"
	"sync"
	"time"
)

const (
	UpdateInterval   = 100 * time.Millisecond
	WarningThreshold = 30000 // 30 seconds in ms
	DangerThreshold  = 10000 // 10 seconds in ms
)

var (
	ticker      *time.Ticker
	stopChan    chan bool
	timerMutex  sync.Mutex
	timerActive bool
)

// Start starts the timer countdown
func Start() {
	timerMutex.Lock()
	defer timerMutex.Unlock()

	if timerActive {
		stopUnsafe()
	}

	stopChan = make(chan bool)
	ticker = time.NewTicker(UpdateInterval)
	timerActive = true

	UpdateDisplay()

	go func() {
		for {
			select {
			case <-ticker.C:
				decrementTime()
				UpdateDisplay()
			case <-stopChan:
				return
			}
		}
	}()
}

// UpdateDisplay updates timer displays
func UpdateDisplay() {
	s := Get()
	times := s.GetTimeRemaining()

	for i := 0; i < 2; i++ {
		timerID := fmt.Sprintf("timer-%d", i)
		ms := times[i]

		// Format time
		timeStr := formatTime(ms)
		SetText(timerID, timeStr)

		// Remove all state classes
		RemoveClass(timerID, "warning")
		RemoveClass(timerID, "danger")

		// Add warning/danger classes
		if ms <= DangerThreshold {
			AddClass(timerID, "danger")
		} else if ms <= WarningThreshold {
			AddClass(timerID, "warning")
		}
	}
}

// Stop stops the timer countdown
func Stop() {
	timerMutex.Lock()
	defer timerMutex.Unlock()
	stopUnsafe()
}

func stopUnsafe() {
	if !timerActive {
		return
	}

	timerActive = false

	if ticker != nil {
		ticker.Stop()
		ticker = nil
	}

	if stopChan != nil {
		close(stopChan)
		stopChan = nil
	}
}

// decrementTime decreases current player's time
func decrementTime() {
	s := Get()
	currentTurn := s.GetCurrentTurn()

	if currentTurn < 0 || currentTurn > 1 {
		return
	}

	times := s.GetTimeRemaining()
	times[currentTurn] -= int64(UpdateInterval / time.Millisecond)

	if times[currentTurn] < 0 {
		times[currentTurn] = 0
	}

	s.SetTimeRemaining(times)
}

// formatTime converts milliseconds to MM:SS
func formatTime(ms int64) string {
	if ms < 0 {
		ms = 0
	}

	totalSeconds := ms / 1000
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

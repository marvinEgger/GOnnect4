// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Astrit Aslani astrit.aslani@gmail.com
// Created: 23.12.2025
//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"

	"github.com/marvinEgger/GOnnect4/client/wasm/lib"
)

// remarshal converts interface{} to struct via JSON
func remarshal(in interface{}, out interface{}) error {
	bytes, err := json.Marshal(in)
	if err != nil {
		lib.Console("Error marshaling in remarshal: " + err.Error())
		return err
	}

	err = json.Unmarshal(bytes, out)
	if err != nil {
		lib.Console("Error unmarshaling in remarshal: " + err.Error())
	}

	return err
}

// formatPlayerCount formats player count message
func formatPlayerCount(count int) string {
	switch count {
	case 0:
		return "0 players online"
	case 1:
		return "1 player online"
	default:
		return fmt.Sprintf("%d players online", count)
	}
}

// getMatchmakingStatus returns matchmaking status text
func getMatchmakingStatus(playerCount int) string {
	switch {
	case playerCount >= 2:
		return "Match found! Starting game..."
	case playerCount == 1:
		return "Waiting for one more player..."
	default:
		return "Looking for available players..."
	}
}

// clearMessage clears a message element
func clearMessage(elementID string) {
	lib.SetText(elementID, "")
	lib.GetElement(elementID).Set("className", "message")
}

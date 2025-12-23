// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author: Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package main

import (
	"encoding/json"

	"github.com/marvinEgger/GOnnect4/server/lib"
)

// mapToStruct converts interface{} to struct via JSON
func mapToStruct(in interface{}, out interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// getPlayerInfos gets public info for both players
func (srv *Server) getPlayerInfos(game *lib.Game) [2]lib.PlayerInfo {
	var infos [2]lib.PlayerInfo
	players := game.GetPlayers()
	for i, p := range players {
		if p != nil {
			infos[i] = lib.PlayerInfo{
				ID:        p.ID,
				Username:  p.Username,
				Connected: p.IsConnected(),
			}
		}
	}
	return infos
}

// getTimeRemaining gets remaining time for both players in milliseconds
func (srv *Server) getTimeRemaining(game *lib.Game) [2]int64 {
	times := game.GetTimeRemaining()
	return [2]int64{
		times[0].Milliseconds(),
		times[1].Milliseconds(),
	}
}

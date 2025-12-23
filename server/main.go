// Copyright (c) 2025 Haute école d'ingénierie et d'architecture de Fribourg
// SPDX-License-Identifier: Apache-2.0
// Author:  Marvin Egger marvin.egger@hotmail.ch
// Created: 05.12.2025

package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	listenAddress = ":8080"
	webFolder     = "./client"
)

// main entry point
func main() {
	// Create and start server
	server := NewServer()
	server.StartPeriodicCleanup()

	// Register the web socket handler
	http.HandleFunc("/ws", server.handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir(webFolder)))

	fmt.Printf("Server starting on %s\n", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

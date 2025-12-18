#!/bin/bash

# Script to run the server
# If already running, kill with : lsof -ti:8080 | xargs kill -9 2>/dev/null; echo "Port 8080 freed"

echo "Starting GOnnect4 server..."
echo "Working directory: $(pwd)"
echo "Server will be available at http://localhost:8080"
echo ""

# Check if server binary exists
if [ ! -f "server/dist/game" ]; then
    echo "Server binary not found. Please run ./build.sh first"
    exit 1
fi

# Check if client folder exists
if [ ! -d "client" ]; then
    echo "Client folder not found. Please ensure you're in the project root"
    exit 1
fi

# Run server
./server/dist/game

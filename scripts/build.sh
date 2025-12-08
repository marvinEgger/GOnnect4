#!/bin/bash

# Script to build server and WASM client

echo "Building GOnnect4..."
echo "Working directory: $(pwd)"
echo ""

# Create dist directory if it doesn't exist
mkdir -p server/dist

# Build server
echo "Building server..."
cd server
go build -o ./dist/game .
SERVER_STATUS=$?
cd ..

if [ $SERVER_STATUS -ne 0 ]; then
    echo "Server build failed"
    exit 1
fi

echo "Server build successful"
echo ""

# Create dist directory if it doesn't exist
mkdir -p client/dist

# Build WASM client
echo "Building WASM client..."
./scripts/build-wasm.sh
WASM_STATUS=$?

if [ $WASM_STATUS -ne 0 ]; then
    echo "WASM build failed"
    exit 1
fi

echo ""
echo "Build completed successfully"
echo "Server binary: server/dist/game"
echo "WASM binary: client/dist/game.wasm"

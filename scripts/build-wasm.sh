#!/bin/bash

echo "Building WASM client..."
cd client/wasm

GOOS=js GOARCH=wasm go build -o ../dist/game.wasm .
BUILD_STATUS=$?

if [ $BUILD_STATUS -ne 0 ]; then
    echo "WASM build failed"
    exit 1
fi

# Copy wasm_exec.js
WASM_EXEC_PATH="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [ ! -f "$WASM_EXEC_PATH" ]; then
    WASM_EXEC_PATH="$(go env GOROOT)/misc/wasm/wasm_exec.js"
fi

if [ -f "$WASM_EXEC_PATH" ]; then
    cp "$WASM_EXEC_PATH" ../dist/wasm_exec.js
    echo "WASM build successful ($(ls -lh ../game.wasm | awk '{print $5}'))"
else
    echo "Warning: wasm_exec.js not found"
    echo "WASM build successful ($(ls -lh ../game.wasm | awk '{print $5}'))"
fi

cd ../..

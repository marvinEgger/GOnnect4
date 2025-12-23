#!/usr/bin/env bash
set -euo pipefail

# Repo root = parent directory of this script (scripts/..)
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Building GOnnect4..."
echo "Repo root: $ROOT_DIR"
echo ""

# Safety checks (évite de recréer des dossiers au mauvais endroit)
[[ -d "$ROOT_DIR/server" ]] || { echo "Error: $ROOT_DIR/server not found"; exit 1; }
[[ -d "$ROOT_DIR/client/wasm" ]] || { echo "Error: $ROOT_DIR/client/wasm not found"; exit 1; }

# --------------------
# Build server
# --------------------
echo "Building server..."
mkdir -p "$ROOT_DIR/server/dist"

(
  cd "$ROOT_DIR/server"
  go build -o "./dist/game" .
)

echo "Server build successful"
echo ""

# --------------------
# Build WASM client
# --------------------
echo "Building WASM client..."
mkdir -p "$ROOT_DIR/client/dist"

(
  cd "$ROOT_DIR/client/wasm"
  GOOS=js GOARCH=wasm go build -o "../dist/game.wasm" .
)

# Copy wasm_exec.js
WASM_EXEC_PATH="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [[ ! -f "$WASM_EXEC_PATH" ]]; then
  WASM_EXEC_PATH="$(go env GOROOT)/misc/wasm/wasm_exec.js"
fi

if [[ -f "$WASM_EXEC_PATH" ]]; then
  cp "$WASM_EXEC_PATH" "$ROOT_DIR/client/dist/wasm_exec.js"
  echo "WASM build successful ($(ls -lh "$ROOT_DIR/client/dist/game.wasm" | awk '{print $5}'))"
else
  echo "Warning: wasm_exec.js not found"
  echo "WASM build successful ($(ls -lh "$ROOT_DIR/client/dist/game.wasm" | awk '{print $5}'))"
fi

echo ""
echo "Build completed successfully"
echo "Server binary: server/dist/game"
echo "WASM binary: client/dist/game.wasm"

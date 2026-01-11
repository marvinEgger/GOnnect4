#!/bin/bash
# Script to lint all Go code in the project

set -euo pipefail

# Get the project root (parent directory of scripts/)
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "🔍 Linting Server..."
cd "$ROOT_DIR/server"
golangci-lint run
echo "✅ Server linting passed"

echo ""
echo "🔍 Linting Client WASM..."
cd "$ROOT_DIR/client/wasm"
GOOS=js GOARCH=wasm golangci-lint run
echo "✅ Client WASM linting passed"

echo ""
echo "✅ All linting passed!"

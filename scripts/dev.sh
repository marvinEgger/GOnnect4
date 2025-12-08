#!/bin/bash
# Require fswatch (optional for the safe mode) : brew install fswatch

# Load .env if exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

PORT=${SERVER_PORT:-8080}

echo "Starting development mode..."
echo "Server will run at http://localhost:$PORT"
echo ""

# Build once first
./scripts/build.sh || exit 1

# Check if fswatch is installed
if command -v fswatch &> /dev/null; then
    echo "Watching for changes (fswatch)..."
    echo "Press Ctrl+C to stop"

    # Start server in background
    ./server/dist/game &
    SERVER_PID=$!

    # Watch and rebuild on changes
    fswatch -o client/*.html client/*.css client/wasm/**/*.go server/**/*.go | while read; do
        echo ""
        echo "Changes detected, rebuilding..."
        kill $SERVER_PID 2>/dev/null
        ./scripts/build.sh && ./server/dist/game &
        SERVER_PID=$!
    done
else
    echo "fswatch not found. Install with: brew install fswatch"
    echo "Running server without watch mode..."
    ./server/dist/game
fi
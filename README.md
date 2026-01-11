# GOnnect4

Online multiplayer Connect 4 game.

Client-Server architecture with a server backend in Go and communication via WebSocket in a WebAssembly client (WASM) for the frontend.

Let's try it on [gonnect4.ch](https://gonnect4.ch/)

## Description

GOnnect4 is an online Connect 4 game allowing two players to compete in real-time.

**Architecture**:
- **Server**: Go + WebSocket for real-time communication
- **Client**: WebAssembly (Go compiled to WASM) + HTML/CSS
- **Communication**: Bidirectional WebSocket
- **Deployment**: Docker + GitHub Container Registry

---

## Local Development

### Prerequisites

- Docker and Docker Compose installed
- GitHub Container Registry authentication (only for private repos)

### Start the application

```bash
# Start (uses docker-compose.dev.yml)
docker compose -f docker-compose.dev.yml up -d --build

# View logs
docker compose -f docker-compose.dev.yml logs -f

# Access the application
open http://localhost:8080
```

### Stop the application

```bash
docker compose -f docker-compose.dev.yml down
```

### Update to latest version

```bash
docker compose -f docker-compose.dev.yml pull && docker compose -f docker-compose.dev.yml up -d
```

---

## Linting

Ensures code quality and consistency across the codebase.

**Prerequisites**: Install `golangci-lint` ([installation guide](https://golangci-lint.run/docs/welcome/quick-start/))
```bash
# macOS/Linux
brew install golangci-lint

# Or using Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Run linter**:
```bash
# Lint Go code (server + WASM client)
./scripts/lint.sh
```

The script runs `golangci-lint` for:
- Server (`server/`)
- WASM client (`client/wasm/`) with `GOOS=js GOARCH=wasm`

---

## Tests

Validates logic and ensures code reliability.

```bash
# Run unit tests
cd server
go test ./...
```

---

## Project Structure

```
GOnnect4/
├── server/                      # Go backend (WebSocket server)
│   ├── main.go                  # Entry point
│   ├── server.go                # Server configuration
│   ├── handlers.go              # WebSocket message handlers
│   ├── websocket.go             # WebSocket connection management
│   ├── matchmaking.go           # Player matchmaking system
│   ├── helpers.go               # Utility functions
│   ├── dist/                    # Compiled binaries
│   └── lib/
│       ├── game.go              # Game state & logic
│       ├── board.go             # Board representation & win detection
│       ├── player.go            # Player representation
│       ├── client.go            # Client connection wrapper
│       ├── protocol.go          # WebSocket protocol messages
│       ├── node.go              # Matchmaking graph node
│       ├── direction.go         # Win-checking directions
│       ├── errors.go            # Manage errors
│       └── *_test.go            # Unit tests
│
├── client/                      # Frontend (WASM + HTML/CSS)
│   ├── index.html               # Main HTML
│   ├── style.css                # Styling
│   ├── assets/                  # Static files (images, etc.)
│   ├── dist/                    # Compiled WASM + wasm_exec.js
│   └── wasm/                    # Go WASM client
│       ├── main.go              # WASM entry point
│       ├── handlers.go          # DOM event handlers
│       ├── helpers.go           # Utility functions
│       └── lib/
│           ├── state.go         # Application state
│           ├── board.go         # Board rendering (Canvas API)
│           ├── client.go        # Server connection
│           ├── dom.go           # DOM manipulation
│           └── timer.go         # Game timer
│
├── scripts/
│   └── lint.sh                  # Linting script
│
├── .github/workflows/
│   └── ci-cd.yml                # CI/CD pipeline
│
├── Dockerfile                   # Used by CI to build images
├── docker-compose.yml           # Production configuration
├── docker-compose.dev.yml       # Development configuration
└── .golangci.yml                # Linting configuration
```

**Architecture principles**:
- **Entry points** (`main.go`): Initialize and start the application
- **Handlers & Helpers**: Simplify the main logic (HTTP/WebSocket handling, utilities)
- **lib/ folders**: Contain parts of logic (game rules, board, state management, etc.)

---

## Docker Compose

The project uses two distinct Docker configurations :

- `docker-compose.dev.yml` (Development)
  - For local development
  - Direct access to `http://localhost:8080`
  - No Nginx needed

- `docker-compose.yml` (Production)
  - For VPS deployment
  - Port **not exposed** publicly (security)
  - Requires Nginx as reverse proxy
  - Static files served by Nginx

---

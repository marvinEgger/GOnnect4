# ==========================================
# STAGE 1: Build the Server and WASM Client
# ==========================================
FROM golang:1.23-alpine AS builder

# Install build tools (git is often needed for go mod download)
RUN apk add --no-cache git

WORKDIR /app

# 1. Dependency Caching Layer
# Copy only the dependency files first. This keeps the build fast
# by not re-downloading modules if only your code changes.
COPY server/go.mod server/go.sum ./server/
COPY client/wasm/go.mod ./client/wasm/

# Download dependencies for Server
WORKDIR /app/server
RUN go mod download

# Download dependencies for WASM Client
WORKDIR /app/client/wasm
RUN go mod download

# 2. Copy the full Source Code
WORKDIR /app
COPY . .

# 3. Build the Go Server
WORKDIR /app/server
RUN mkdir -p dist
RUN go build -o dist/game .

# 4. Build the WASM Client
# Corresponds to your build.sh: GOOS=js GOARCH=wasm go build ...
WORKDIR /app/client/wasm
ENV GOOS=js
ENV GOARCH=wasm
RUN go build -o ../dist/game.wasm .

# 5. Get the wasm_exec.js glue code
# We find it in the GOROOT (standard Go library location)
RUN cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ../dist/wasm_exec.js

# ==========================================
# STAGE 2: Create the Final Runtime Image
# ==========================================
FROM alpine:latest

WORKDIR /app

# Copy the compiled Server binary from the builder
COPY --from=builder /app/server/dist/game ./server/dist/game

# Copy the Client assets (HTML, CSS, plus the WASM/JS we just built)
# This includes the "client" folder structure so your server can serve "./client/..."
COPY --from=builder /app/client ./client

# Expose the port your server listens on (matches run.sh)
EXPOSE 8080

CMD ["./server/dist/game"]
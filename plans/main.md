# Port Forwarding Tunnel — Full Plan

## Overview

A self-hostable HTTP tunneling system (similar to ngrok) that exposes a local service to the internet via a public server. Built in Go with WebSocket as the tunnel transport.

One server ↔ One client. The client connects to the server, and all incoming HTTP traffic is forwarded through the WebSocket tunnel to the client, which replays it against a local service and sends the response back.

---

## Repo Structure

```
tunnel/
├── go.mod
├── go.sum
├── cmd/
│   ├── server/
│   │   └── main.go          # Server entrypoint
│   └── client/
│       └── main.go          # Client entrypoint
├── internal/
│   ├── proto/
│   │   └── message.go       # Shared WS message types
│   ├── server/
│   │   ├── server.go        # HTTP server + WS upgrade
│   │   ├── tunnel.go        # WS tunnel manager
│   │   └── proxy.go         # HTTP handler → tunnel bridge
│   └── client/
│       ├── client.go        # WS connection to server
│       └── forwarder.go     # Replays request to localhost
└── README.md
```

Single `go.mod` at root covering both binaries.

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/gorilla/websocket` | WebSocket server + client (industry standard) |
| Standard library only for everything else | `net/http`, `encoding/json`, `context`, `sync`, `log` |

---

## Protocol — WS Message Format

All messages over the WebSocket tunnel are JSON.

### Request Message (Server → Client)

```
{
  "id":      "uuid-v4",         // unique request ID
  "type":    "request",
  "method":  "GET",
  "path":    "/api/something",
  "headers": { "Content-Type": "application/json", ... },
  "body":    "base64-encoded-body-or-empty-string"
}
```

### Response Message (Client → Server)

```
{
  "id":      "uuid-v4",         // same ID as the request
  "type":    "response",
  "status":  200,
  "headers": { "Content-Type": "application/json", ... },
  "body":    "base64-encoded-body-or-empty-string"
}
```

Body is base64-encoded to safely carry binary payloads through JSON.

---

## Authentication

A shared secret is used to authenticate the client during the WebSocket handshake.

- The client sends the secret as a custom HTTP header during the upgrade request: `X-Tunnel-Secret: <secret>`
- The server checks this header before completing the upgrade
- If missing or wrong → `401` and connection closed
- Secret is configured via CLI flag on both sides (`--secret`)

No token rotation, no expiry — kept intentionally simple. The operator is responsible for choosing a strong secret.

---

## Server — How It Works

### Responsibilities
1. Accept incoming HTTP requests from the public internet on one port (e.g. `:8080`)
2. Accept exactly one WebSocket client connection on a separate path (e.g. `/tunnel`)
3. For each HTTP request:
   - Generate a unique request ID
   - Serialize request into a WS message
   - Register a pending channel keyed by request ID
   - Send the message over the WebSocket
   - Block and wait for the response on that channel (with timeout)
   - Write the response back to the HTTP caller
4. When a WS response message arrives, look up the channel by ID and send the response into it

### Concurrency model
- Each incoming HTTP request is handled in its own goroutine (standard Go HTTP)
- A single goroutine owns the WS write loop (gorilla/websocket is not concurrent-write-safe, so writes must be serialized)
- A separate goroutine owns the WS read loop, dispatching responses to pending channels
- Pending requests stored in a `sync.Map` keyed by request ID

### Only one client
- If a client is already connected and another tries to connect → reject with `409`
- On client disconnect, the server clears the active connection and returns `502` to any in-flight requests

### Timeout
- Each HTTP request waiting for a tunnel response has a configurable timeout (default: 30s)
- On timeout → `504 Gateway Timeout`

---

## Client — How It Works

### Responsibilities
1. Connect to the server's `/tunnel` WebSocket endpoint with the shared secret header
2. Maintain the connection (reconnect on disconnect with exponential backoff)
3. Read request messages from the WS
4. For each request, spin up a goroutine that:
   - Reconstructs the HTTP request
   - Sends it to `localhost:<port>`
   - Reads the response
   - Serializes it into a WS response message
   - Sends it back over the WS
5. Serialize all WS writes through a single write goroutine (same reason as server)

### CLI flags

| Flag | Default | Description |
|---|---|---|
| `--server` | — | WebSocket server URL (e.g. `wss://myserver.onrender.com/tunnel`) |
| `--port` | `3000` | Local port to forward to |
| `--secret` | — | Shared secret for auth |

### Reconnect strategy
- On disconnect, wait and retry: 1s → 2s → 4s → 8s → max 30s
- Log each reconnect attempt
- No limit on retry count — it keeps trying until killed

---

## Server — CLI flags

| Flag | Default | Description |
|---|---|---|
| `--port` | `8080` | Port to listen on for HTTP traffic |
| `--secret` | — | Shared secret to authenticate the client |

---

## Request/Response Flow (Step by Step)

```
1. Browser sends GET /api/hello to server:8080

2. Server HTTP handler fires in goroutine:
   - Generates request ID: "abc-123"
   - Creates pending channel: pendingMap["abc-123"] = make(chan Response, 1)
   - Encodes request as JSON WS message
   - Acquires WS write lock, sends message to client

3. Client WS read loop receives the message:
   - Spawns goroutine to handle it
   - Reconstructs HTTP request
   - Calls localhost:3000/api/hello
   - Gets response back
   - Encodes response as JSON WS message with ID "abc-123"
   - Acquires WS write lock, sends response back to server

4. Server WS read loop receives response:
   - Looks up pendingMap["abc-123"]
   - Sends response into the channel
   - Deletes entry from map

5. Server HTTP handler unblocks:
   - Reads response from channel
   - Writes status, headers, body back to browser
```

---

## Error Cases

| Scenario | Behavior |
|---|---|
| No client connected | `502 Bad Gateway` |
| Client disconnects mid-request | Pending requests receive error, respond with `502` |
| Local service unreachable | Client sends back `502` response through tunnel |
| Request timeout (30s) | `504 Gateway Timeout` |
| Wrong secret on WS upgrade | `401 Unauthorized`, connection closed |
| Second client tries to connect | `409 Conflict`, connection rejected |
| Local service returns any status | Passed through as-is |

---

## What's Out of Scope (for now)

- HTTPS termination on the server (handled by the hosting platform e.g. Render)
- WebSocket proxying (HTTP only, as planned)
- Multiple clients / multiple ports
- A dashboard or web UI
- Request logging beyond stdout

---

## Deployment Notes (Render)

- Deploy `cmd/server` as a Web Service on Render
- Render provides HTTPS + a public URL automatically
- Set `--secret` as an environment variable on Render
- The client points `--server` to the Render-provided `wss://` URL
- Render's free tier keeps the service alive as long as traffic comes in; the client's reconnect logic handles any cold starts

---

## Build Commands (planned)

```
# Build server
go build -o bin/server ./cmd/server

# Build client
go build -o bin/client ./cmd/client

# Run server
./bin/server --port 8080 --secret mysecret

# Run client
./bin/client --server wss://myapp.onrender.com/tunnel --port 3000 --secret mysecret
```

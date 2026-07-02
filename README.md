# go-tunnel

A self-hostable HTTP tunneling system (similar to ngrok) that exposes a local service to the internet via a public server. Built in Go with WebSocket as the tunnel transport.

One server ↔ One client. The client connects to the server, and all incoming HTTP traffic is forwarded through the WebSocket tunnel to the client, which replays it against a local service and sends the response back.

---

## How It Works

```
Browser ──HTTP──▶ Server (:8080) ──WS──▶ Client ──HTTP──▶ localhost:3000
                        │                                    │
                        ◄────────────────────────────────────┘
                        ◄─────────── WS ─────────────────────┘
                        ◄───────── HTTP ──────────────────────┘
```

1. A browser sends an HTTP request to the public server.
2. The server serializes the request into a JSON message and sends it over the WebSocket tunnel.
3. The client receives the message, reconstructs the HTTP request, and forwards it to a local service.
4. The local service responds; the client serializes the response and sends it back over the WebSocket.
5. The server reads the response and writes it back to the browser.

## Repo Structure

```
├── cmd/
│   ├── server/
│   │   └── main.go              # Server entrypoint
│   └── client/
│       └── main.go              # Client entrypoint
├── internal/
│   ├── protocol/
│   │   ├── doc.go               # Package documentation
│   │   ├── types.go             # Request/Response message types
│   │   └── json.go              # Custom JSON marshal/unmarshal
│   ├── server/
│   │   ├── server.go            # HTTP server + WebSocket upgrade
│   │   ├── tunnel.go            # WebSocket tunnel manager
│   │   └── proxy.go             # HTTP handler → tunnel bridge
│   ├── client/
│   │   ├── cli_flag.go          # CLI flag parsing
│   │   ├── client.go            # WebSocket client + reconnect logic
│   │   └── forwarder.go         # Replays requests to localhost
│   └── logging/
│       └── logging.go           # Simple structured logger
├── go.mod
├── go.sum
└── README.md
```

## Protocol

All messages over the WebSocket tunnel are JSON with a `type` field distinguishing direction.

**Request (Server → Client):**

```json
{
  "id":      "uuid-v4",
  "type":    "request",
  "method":  "GET",
  "path":    "/api/hello",
  "headers": { "Content-Type": "application/json" },
  "body":    "base64-encoded-body"
}
```

**Response (Client → Server):**

```json
{
  "id":      "uuid-v4",
  "type":    "response",
  "status":  200,
  "headers": { "Content-Type": "application/json" },
  "body":    "base64-encoded-body"
}
```

The `Body` field is `[]byte` — Go's `encoding/json` automatically base64-encodes it on marshal and base64-decodes on unmarshal, making the round-trip lossless for any binary or text payload.

## Prerequisites

- Go 1.26+

## Build

```sh
# Build server
go build -o bin/server ./cmd/server

# Build client
go build -o bin/client ./cmd/client
```

## Configuration

### Server

The server is configured via environment variables (loaded from `.env` at startup via `godotenv`).

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Port to listen on for HTTP traffic |
| `SECRET` | — | Shared secret to authenticate the client |

Example `.env`:

```
PORT=8080
SECRET=my-strong-secret
```

### Client

The client uses CLI flags.

| Flag | Default | Description |
|---|---|---|
| `--server` | — | WebSocket server URL — use `ws://` for local testing, `wss://` for production (e.g. `ws://localhost:8080` or `wss://myapp.onrender.com`) |
| `--port` | — | Local port to forward to |
| `--secret` | — | Shared secret for auth |

## Run

### 1. Start a local service to forward to

```sh
python3 -m http.server 3000
```

### 2. Start the server

```sh
# With .env file
./bin/server

# Or with inline env vars
PORT=8080 SECRET=my-strong-secret ./bin/server
```

### 3. Start the client

```sh
./bin/client --server ws://localhost:8080 --port 3000 --secret my-strong-secret
```

### 4. Test it

```sh
curl http://localhost:8080/
```

Requests to the server are forwarded through the tunnel to your local service.

## Authentication

A shared secret authenticates the client during the WebSocket handshake:

- The client sends `X-Tunnel-Secret: <secret>` as an HTTP header during the WebSocket upgrade.
- The server validates the header before completing the upgrade.
- Wrong or missing secret → `401 Unauthorized`.

## Error Handling

| Scenario | Response |
|---|---|
| No client connected | `502 Bad Gateway` |
| Client disconnects mid-request | `502 Bad Gateway` |
| Local service unreachable | Client sends back `502` |
| Request timeout (30s) | `504 Gateway Timeout` |
| Wrong secret on WS upgrade | `401 Unauthorized` |
| Second client attempts connection | `409 Conflict` |

## Deployment (Render)

1. Deploy `cmd/server` as a Web Service on Render.
2. Use the following **Build Command** in the Render dashboard:

   ```sh
   go build -tags netgo -ldflags '-s -w' -o app ./cmd/server
   ```

3. Set `PORT` and `SECRET` as environment variables in the Render dashboard.
4. Render provides HTTPS + a public URL automatically.
5. Point the client to the Render-provided `wss://` URL:

```sh
./bin/client --server wss://myapp.onrender.com --port 3000 --secret my-strong-secret
```

The client's built-in reconnect logic handles any cold starts on Render's free tier.

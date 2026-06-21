# Implementation Steps

## Phase 1 — Project Setup

- [x] Initialize Go module and .gitignore
- [x] Add `gorilla/websocket` dependency
- [x] Create directory structure: `cmd/server/`, `cmd/client/`, `internal/protocol/`, `internal/server/`, `internal/client/`

## Phase 2 — Shared Protocol (`internal/protocol/`)

- [x] Define `Request` and `Response` message structs with JSON tags
- [x] Define message type constants (`MsgTypeRequest`, `MsgTypeResponse`)
- [x] Write helpers: `EncodeMessage`, `DecodeMessage` (custom MarshalJSON/UnmarshalJSON)

## Phase 3 — Server (`cmd/server/` + `internal/server/`)

- [x] CLI flags: `--port`, `--secret`
- [x] WebSocket upgrade handler on `/__tunnel__` with secret auth check
- [x] Tunnel manager: track single active client, reject duplicates (409)
- [x] Proxy handler: capture incoming HTTP, send via WS, wait on channel, respond
- [x] WS read loop: receive responses, dispatch to pending channels
- [x] WS write loop: single goroutine serializing writes
- [x] Timeout handling per request (default 30s → 504)
- [x] Client disconnect → drain/error pending requests (502)

## Phase 4 — Client (`cmd/client/` + `internal/client/`)

- [x] CLI flags: `--server`, `--port`, `--secret`
- [x] Connect to server WS with `X-Tunnel-Secret` header
- [x] WS read loop: receive requests, forward to localhost, send response
- [x] Request handler: reconstruct HTTP, forward to localhost, collect response
- [x] WS write loop: serialize responses back to server
- [ ] Reconnect with exponential backoff (1s → 2s → 4s → 8s → max 30s)

## Phase 5 — Integration & Testing

- [ ] Manual test: server + client + local HTTP service
- [ ] Edge cases: wrong secret, duplicate client, timeout, local service down
- [ ] Build scripts in README or Makefile

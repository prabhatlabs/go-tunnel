# Implementation Steps

## Phase 1 — Project Setup
- [x] Initialize Go module and .gitignore
- [ ] Add `gorilla/websocket` dependency
- [ ] Create directory structure: `cmd/server/`, `cmd/client/`, `internal/proto/`, `internal/server/`, `internal/client/`

## Phase 2 — Shared Protocol (`internal/proto/`)
- [ ] Define `Request` and `Response` message structs with JSON tags
- [ ] Define message type constants (`MsgTypeRequest`, `MsgTypeResponse`)
- [ ] Write helpers: `EncodeMessage`, `DecodeMessage`

## Phase 3 — Server (`cmd/server/` + `internal/server/`)
- [ ] CLI flags: `--port`, `--secret`
- [ ] WebSocket upgrade handler on `/__tunnel__` with secret auth check
- [ ] Tunnel manager: track single active client, reject duplicates (409)
- [ ] Proxy handler: capture incoming HTTP, send via WS, wait on channel, respond
- [ ] WS read loop: receive responses, dispatch to pending channels
- [ ] WS write loop: single goroutine serializing writes
- [ ] Timeout handling per request (default 30s → 504)
- [ ] Client disconnect → drain/error pending requests (502)

## Phase 4 — Client (`cmd/client/` + `internal/client/`)
- [ ] CLI flags: `--server`, `--port`, `--secret`
- [ ] Connect to server WS with `X-Tunnel-Secret` header
- [ ] WS read loop: receive requests, spawn handler goroutines
- [ ] Request handler: reconstruct HTTP, forward to localhost, collect response
- [ ] WS write loop: serialize responses back to server
- [ ] Reconnect with exponential backoff (1s → 2s → 4s → 8s → max 30s)

## Phase 5 — Integration & Testing
- [ ] Manual test: server + client + local HTTP service
- [ ] Edge cases: wrong secret, duplicate client, timeout, local service down
- [ ] Build scripts in README or Makefile

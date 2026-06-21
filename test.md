# Testing

## Running tests

```sh
go test ./internal/protocol/
```

## Round-trip behavior

The `Body` field is `[]byte`, which Go's `encoding/json` automatically
base64-encodes on marshal and base64-decodes on unmarshal.

This means:

1. Server reads raw HTTP body bytes → MarshalJSON base64-encodes them
2. Client UnmarshalJSON base64-decodes → forwards original bytes to local service

The round-trip is **lossless** for any byte sequence — plain text, binary,
or pre-encoded base64 strings. The local service always receives exactly
what the original caller sent.

### What to verify

- `[]byte` body round-trips correctly (binary, text, empty, nil)
- Pre-encoded base64 body is forwarded as-is (no double-encoding)
- `Type` field is enforced ("request" / "response" only)
- Wrong `Type` value on the wire is rejected during unmarshal

## Manual integration test

Start a local HTTP service to forward to:

```sh
python3 -m http.server 3000
```

Build and run the server:

```sh
go build -o bin/server ./cmd/server
./bin/server --port 8080 --secret mysecret
```

In another terminal, run the client:

```sh
go build -o bin/client ./cmd/client
./bin/client --server localhost:8080 --port 3000 --secret mysecret
```

Now hit the server — requests should forward through the tunnel to the local service:

```sh
curl http://localhost:8080/
```

### Error scenarios to test

| Test | Expected |
|---|---|
| No client connected | `502 Bad Gateway` |
| Wrong secret | `401 Unauthorized` |
| Second client | `409 Conflict` |
| Local service down | Client sends back `502` |
| Request timeout (30s) | `504 Gateway Timeout` |

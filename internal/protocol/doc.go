// Package protocol defines the shared WebSocket message types
// used for communication between the tunnel server and client.
//
// Messages are JSON-encoded and distinguished by the "type" field:
// "request" for requests forwarded from server to client,
// "response" for replies sent back from client to server.
package protocol

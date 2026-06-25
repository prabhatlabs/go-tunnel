package client

import (
	"flag"

	"github.com/prabhatlabs/go-tunnel/internal/logging"
)

func GetFlags() (string, int, string) {
	serverUrlPtr := flag.String("server", "", "Server URL (e.g. ws://localhost:8080 or wss://myapp.onrender.com)")
	portPtr := flag.Int("port", 0, "Port on which local service running")
	secretPtr := flag.String("secret", "", "Secret for the server")
	flag.Parse()

	serverUrl := *serverUrlPtr
	port := *portPtr
	secret := *secretPtr

	if serverUrl == "" {
		logging.Error("Server url must be provided")
	}

	if port < 1 || port > 65535 {
		logging.Error("Port must be between 1 and 65535")
	}

	if secret == "" {
		logging.Error("Secret must be provided")
	}

	return serverUrl, port, secret
}

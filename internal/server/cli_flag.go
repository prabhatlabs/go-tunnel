package server

import (
	"flag"

	"github.com/prabhatlabs/go-tunnel/internal/logging"
)

func GetFlags() (int, string) {
	portPtr := flag.Int("port", 8080, "Port for the server (default: 8080)")
	secretPtr := flag.String("secret", "", "Secret for the server")
	flag.Parse()

	port := *portPtr
	secret := *secretPtr

	if port < 1 || port > 65535 {
		logging.Error("Port must be between 1 and 65535")
	}

	if secret == "" {
		logging.Error("Secret must be provided")
	}

	return port, secret
}

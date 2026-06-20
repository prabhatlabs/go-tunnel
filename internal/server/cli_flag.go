package server

import (
	"flag"
	"log"
)

func GetFlags() (int, string) {
	portPtr := flag.Int("port", 8080, "Port for the server (default: 8080)")
	secretPtr := flag.String("secret", "", "Secret for the server")
	flag.Parse()

	port := *portPtr
	secret := *secretPtr

	if port < 1 || port > 65535 {
		log.Fatalln("Port must be between 1 and 65535")
	}

	if secret == "" {
		log.Fatalln("Secret must be provided")
	}

	return port, secret
}

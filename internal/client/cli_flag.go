package client

import (
	"flag"
	"log"
)

func GetFlags() (string, int, string) {
	serverUrlPtr := flag.String("server", "", "Server url")
	portPtr := flag.Int("port", 0, "Port on which local service running")
	secretPtr := flag.String("secret", "", "Secret for the server")
	flag.Parse()

	serverUrl := *serverUrlPtr
	port := *portPtr
	secret := *secretPtr

	if serverUrl == "" {
		log.Fatalln("Server url must be provided")
	}

	if port < 1 || port > 65535 {
		log.Fatalln("Port must be between 1 and 65535")
	}

	if secret == "" {
		log.Fatalln("Secret must be provided")
	}

	return serverUrl, port, secret
}

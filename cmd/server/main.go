package main

import (
	"log"

	"github.com/prabhatlabs/go-tunnel/internal/server"
)

func main() {
	port, secret := server.GetFlags()

	srv := server.New(port, secret)
	log.Printf("listening on :%d", port)
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}

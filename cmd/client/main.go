package main

import (
	"log"

	"github.com/prabhatlabs/go-tunnel/internal/client"
)

func main() {
	url, port, secret := client.GetFlags()
	c := client.New(url, port, secret)

	log.Printf("forwarding port :%d", port)

	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}

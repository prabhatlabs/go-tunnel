package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/prabhatlabs/go-tunnel/internal/logging"
	"github.com/prabhatlabs/go-tunnel/internal/server"
)

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	secret := os.Getenv("SECRET")
	srv := server.New(port, secret)

	if port == "" {
		logging.Error("Add PORT in .env")
	}
	logging.Info("Listening on :", port)

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}

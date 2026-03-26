package main

import (
	"log"

	"fleexa/internal/api"
)

func main() {
	server := api.NewServer()

	if err := server.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
package main

import (
	"log"

	"github.com/Fleexa-Graduation-Project/Backend/internal/api"
)

func main() {
	server := api.NewServer()

	if err := server.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
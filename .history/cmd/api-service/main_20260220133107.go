package main

import (
	"net/http"
	"os"

	"github.com/Fleexa-Graduation-Project/Backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Set up our logger (the exact same one we used in the Ingestion Lambda!)
	log := logger.InitLogger()
	log.Info("Starting Fleexa API Server...")

	// 2. Create the Gin "Engine"
	// gin.Default() automatically adds crash-protection and logs every request to the terminal
	router := gin.Default()

	// 3. Define our very first Route (The Heartbeat)
	router.GET("/ping", func(c *gin.Context) {
		// c.JSON formats our response perfectly for the Flutter app
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "Fleexa API is alive! 🚀",
		})
	})

	// 4. Figure out which port to listen on (Default to 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 5. Open the restaurant doors! (Start the server)
	log.Info("Listening and serving HTTP on port: " + port)
	if err := router.Run(":" + port); err != nil {
		log.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
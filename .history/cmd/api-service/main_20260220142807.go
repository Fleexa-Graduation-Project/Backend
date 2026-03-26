package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Fleexa-Graduation-Project/Backend/internal/api/handlers" // Import your new handler!
	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/db"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	log := logger.InitLogger()
	log.Info("Starting Fleexa API Server...")

	// 1. Initialize DB
	if err := db.NewDynamoDBClient(context.Background()); err != nil {
		log.Error("Failed to initialize DynamoDB", "error", err)
		panic(err)
	}

	// 2. Initialize the StateStore (The Chef)
	stateStore, err := devices.NewStateStore()
	if err != nil {
		log.Error("Failed to initialize StateStore", "error", err)
		panic(err)
	}

	// 3. Initialize the DeviceHandler (The Waiter)
	deviceHandler := &handlers.DeviceHandler{
		StateStore: stateStore,
	}

	router := gin.Default()

	// --- THE MENU (Routes) ---
	
	// Our old ping test
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// NEW: The live devices route!
	router.GET("/devices", deviceHandler.GetDevices)

	// --------------------------

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("Listening and serving HTTP on port: " + port)
	if err := router.Run(":" + port); err != nil {
		log.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
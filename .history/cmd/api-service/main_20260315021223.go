package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Fleexa-Graduation-Project/Backend/internal/api/handlers"
	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/internal/telemetry"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/db"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/logger"
	"github.com/Fleexa-Graduation-Project/Backend/internal/alerts"
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

	// 2. Initialize the Stores
	stateStore, err := devices.NewStateStore()
	if err != nil {
		log.Error("Failed to initialize StateStore", "error", err)
		panic(err)
	}

	telemetryStore, err := telemetry.NewTelemetryStore()
	if err != nil {
		log.Error("Failed to initialize TelemetryStore", "error", err)
		panic(err)
	}

	alertStore, err := alerts.NewAlertStore()
if err != nil {
    log.Error("Failed to initialize AlertStore", "error", err)
    panic(err)
}

	// 3. Initialize the DeviceHandler
	deviceHandler := &handlers.DeviceHandler{
		StateStore:     stateStore,
		TelemetryStore: telemetryStore,
		AlertStore:     alertStore,
	}

	router := gin.Default()

	// --- THE MENU (Routes) ---
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Grouping routes under /api/v1 just like in your api_spec.md
	v1 := router.Group("/api/v1")
	{
		v1.GET("/devices", deviceHandler.GetDevices)
		v1.GET("/devices/:id", deviceHandler.GetDeviceByID)
		v1.GET("/devices/:id/telemetry", deviceHandler.GetDeviceTelemetry)
		v1.GET("/devices/:id/alerts", deviceHandler.GetDeviceAlerts)
	}
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
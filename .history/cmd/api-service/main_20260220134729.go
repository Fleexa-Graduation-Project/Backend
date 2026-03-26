package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Fleexa-Graduation-Project/Backend/pkg/db"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	log := logger.InitLogger()
	log.Info("Starting Fleexa API Server...")

	// ---------------------------------------------------------
	// NEW: Connect to the Fridge (DynamoDB) before opening!
	// ---------------------------------------------------------
	if err := db.NewDynamoDBClient(context.Background()); err != nil {
		log.Error("Failed to initialize DynamoDB", "error", err)
		panic(err) // If we can't connect to the DB, the server shouldn't start!
	}
	log.Info("Successfully connected to DynamoDB!")

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "Fleexa API is alive and connected to DB! 🚀",
		})
	})

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
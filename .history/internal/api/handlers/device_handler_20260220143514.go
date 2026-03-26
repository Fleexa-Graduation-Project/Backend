package handlers

import (
	"log" // Added this to see the secret error in the terminal
	"net/http"

	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	StateStore *devices.StateStore
}

// GetDevices returns the live status of all IoT devices
func (h *DeviceHandler) GetDevices(c *gin.Context) {
	// 1. Tell the Chef to grab all states from the fridge
	states, err := h.StateStore.GetAllStates(c.Request.Context())
	
	if err != nil {
		// 🛑 THIS PRINTS THE REAL ERROR TO YOUR TERMINAL
		log.Printf("CRITICAL DATABASE ERROR: %v", err)

		// If something went wrong in the kitchen, tell the customer
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch device states",
		})
		return
	}

	// 2. Serve the data nicely on a plate (JSON)
	c.JSON(http.StatusOK, gin.H{
		"data": states,
	})
}
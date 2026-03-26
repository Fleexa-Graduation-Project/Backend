package handlers

import (
	"net/http"
	"strconv"

	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/internal/telemetry"
	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	StateStore     *devices.StateStore
	TelemetryStore *telemetry.TelemetryStore // The handler can now talk to the Telemetry DB!
}

// GetDevices handles GET /devices
func (h *DeviceHandler) GetDevices(c *gin.Context) {
	states, err := h.StateStore.GetAllStates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch device states"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": states})
}

// GetDeviceByID handles GET /devices/:id
func (h *DeviceHandler) GetDeviceByID(c *gin.Context) {
	deviceID := c.Param("id") // Gin extracts the ID from the URL automatically!

	state, err := h.StateStore.GetStateByID(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if state == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		if err
		return
	}

	c.JSON(http.StatusOK, state)
}

// GetDeviceTelemetry handles GET /devices/:id/telemetry
func (h *DeviceHandler) GetDeviceTelemetry(c *gin.Context) {
	deviceID := c.Param("id")

	// Look for an optional ?limit=X parameter (default to 50 so we don't crash the app)
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	history, err := h.TelemetryStore.GetTelemetryHistory(c.Request.Context(), deviceID, int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch telemetry history"})
		return
	}

	// Format exactly how your api_spec.md promised!
	c.JSON(http.StatusOK, gin.H{
		"device_id": deviceID,
		"data":      history,
	})
}
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
func (handler *DeviceHandler) GetDeviceByID(context *gin.Context) {
	deviceID := context.Param("id") // Gin extracts the ID from the URL automatically!

	state, err := handler.StateStore.GetStateByID(context.Request.Context(), deviceID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if state == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	context.JSON(http.StatusOK, state)
}

// GetDeviceTelemetry handles GET /devices/:id/telemetry
func (handler *DeviceHandler) GetDeviceTelemetry(context *gin.Context) {
	deviceID := context.Param("id")

	// Look for an optional ?limit=X parameter (default to 50 so we don't crash the app)
	limitStr := context.DefaultQuery("limit", "50")
	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	history, err := handler.TelemetryStore.GetTelemetryHistory(context.Request.Context(), deviceID, int32(limit))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch telemetry history"})
		return
	}

	// Format exactly how your api_spec.md promised!
	context.JSON(http.StatusOK, gin.H{
		"device_id": deviceID,
		"data":      history,
	})
}


func isHotTier(period string) bool {
	switch period {
	case "1h", "24h", "5d", "7d":
		return true
	default:	
		return false  //it was deleted by ttl and stored in S3
	}
}
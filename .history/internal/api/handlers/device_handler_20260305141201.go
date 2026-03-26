package handlers

import (
	"net/http"

	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/internal/telemetry"
	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	StateStore     *devices.StateStore
	TelemetryStore *telemetry.TelemetryStore
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
	deviceID := context.Param("id")

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

// GetDeviceTelemetry handles GET /devices/:id/telemetry?period=...&metric=...
func (handler *DeviceHandler) GetDeviceTelemetry(context *gin.Context) {
	deviceID := context.Param("id")
	period := context.DefaultQuery("period", "24h")
	metric := context.DefaultQuery("metric", "temp") // UI tells us what to chart (temp, power, gas_level)

	if isHotTier(period) {
		// Route A: Hot Tier (DynamoDB)
		rawData, dbErr := handler.TelemetryStore.GetTelemetryHistory(context.Request.Context(), deviceID, 0)
		if dbErr != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent history"})
			return
		}

		chartData := telemetry.AggregateTimeSeries(rawData, metric, period)

		context.JSON(http.StatusOK, gin.H{
			"device_id": deviceID,
			"period":    period,
			"source":    "DynamoDB",
			"data":      chartData,
		})

	} else {
		// Route B: Cold Tier (AWS S3) - Placeholder
		context.JSON(http.StatusOK, gin.H{
			"device_id": deviceID,
			"period":    period,
			"source":    "S3",
			"data":      []telemetry.ChartPoint{}, // Empty array for now until we build S3
		})
	}
}

func isHotTier(period string) bool {
	switch period {
	case "1h", "24h", "5d", "7d":
		return true
	default:
		return false
	}
}
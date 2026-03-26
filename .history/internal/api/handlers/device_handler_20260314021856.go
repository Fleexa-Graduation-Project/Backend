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

//handling GET /devices
func (h *DeviceHandler) GetDevices(c *gin.Context) {
	states, err := h.StateStore.GetAllStates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch device states"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": states})
}

//handling GET /devices/:id
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

//handling GET /devices/:id/telemetry?period=...&metric=...
func (handler *DeviceHandler) GetDeviceTelemetry(context *gin.Context) {
	deviceID := context.Param("id")
	period := context.DefaultQuery("period", "24h")
	metric := context.DefaultQuery("metric", "temp") 
	
	response := gin.H{
		"device_id": deviceID,
		"period":    period,
	}

	if isHotTier(period) {
		rawData, dbErr := handler.TelemetryStore.GetTelemetryHistory(context.Request.Context(), deviceID, 0)
		if dbErr != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent history"})
			return
		}

		response["source"] = "DynamoDB"
		response["data"] = telemetry.FilterTime(rawData, metric, period)

		if deviceID == "temp-sensor-01" {
			// We already have the DynamoDB data, just pass it to our 24h strict math function
			stats, _ := telemetry.Calculate24hStats(rawData, metric)
			response["min"] = stats.Min
			response["max"] = stats.Max
			response["average"] = stats.Average
		}

	} else {
		response["source"] = "S3"
		response["data"] = []telemetry.ChartPoint{} // will handle s3 later

	
		if deviceID == "temp-sensor-01" {
			// Even if the chart uses S3 data (e.g. 1 Month), the Stat Blocks STILL need the last 24h!
			// So we do a quick explicit fetch from DynamoDB just for the stats.
			recentData, _ := handler.TelemetryStore.GetTelemetryHistory(context.Request.Context(), deviceID, 0)
			stats, _ := telemetry.Calculate24hStats(recentData, metric)
			response["min"] = stats.Min
			response["max"] = stats.Max
			response["average"] = stats.Average
		}
	}

	// Send the final JSON to Flutter
	context.JSON(http.StatusOK, response)

		chartData := telemetry.FilterTime(rawData, metric, period)

		context.JSON(http.StatusOK, gin.H{
			"device_id": deviceID,
			"period":    period,
			"source":    "DynamoDB",
			"data":      chartData,
		})

	} else {
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
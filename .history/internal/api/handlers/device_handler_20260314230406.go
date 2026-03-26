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

	if state.Type == "light-sensor" {
		if luxVal, exists := state.Payload["light_level"]; exists {
			var lux float64
			
			if floatVal, ok := luxVal.(float64); 
			ok {
				lux = floatVal
			} else if intVal, ok := luxVal.(int); 
			ok {
				lux = float64(intVal)
			}
			
			//calculating status based on lux thresholds
			status := "Bright"
			if lux <= 50 {
				status = "Dark"
			} else if lux <= 500 {
				status = "Normal"
			}
			
			// Inject it directly into the payload for Flutter to read
			state.Payload["light_status"] = status
		}
	}

	context.JSON(http.StatusOK, state)
}

//handling GET /devices/:id/telemetry?period=...&metric=...
func (handler *DeviceHandler) GetDeviceTelemetry(context *gin.Context) {
	deviceID := context.Param("id")
	period := context.DefaultQuery("period", "24h")
	metric := context.DefaultQuery("metric", "temp") 
	
    state, err := handler.StateStore.GetStateByID(context.Request.Context(), deviceID)
    if err != nil {
        context.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }
    if state == nil {
        context.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
        return
    }
	
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

		if state.Type == "temp-sensor" {
			stats, _ := telemetry.CalculateTempState(rawData, metric)
			response["min"] = stats.Min
			response["max"] = stats.Max
			response["average"] = stats.Average
		}

	} else {
		response["source"] = "S3"
		response["data"] = []telemetry.ChartPoint{} // will handle s3 later

		if state.Type == "temp-sensor" {
			recentData, _ := handler.TelemetryStore.GetTelemetryHistory(context.Request.Context(), deviceID, 0)
			stats, _ := telemetry.CalculateTempState(recentData, metric)
			response["min"] = stats.Min
			response["max"] = stats.Max
			response["average"] = stats.Average
		}
	}

	context.JSON(http.StatusOK, response)
} 

func isHotTier(period string) bool {
	switch period {
	case "1h", "24h", "5d", "7d":
		return true
	default:
		return false
	}
}
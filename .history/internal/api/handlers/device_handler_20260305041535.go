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

//handles GET /devices/:id/telemetry?period=
func (handler *DeviceHandler) GetDeviceTelemetry(context *gin.Context) {
	deviceID := context.Param("id")
	period := context.DefaultQuery("period", "24h") 

	var responseData interface{}

	if isHotTier(period) {
		rawData, dbErr := handler.TelemetryStore.GetTelemetryHistory(context.Request.Context(), deviceID, 0)
		if dbErr != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent history"})
			return
		}
		
		responseData = rawData 

	} else {
		// Route B: The data is old. Ask the Librarian (AWS S3)
		// We haven't built the S3 store yet, so we will return a temporary mock response
		
		// s3Data, s3Err := h.S3Store.GetAggregatedHistory(ctx, deviceID, period)
		// responseData = s3Data
		
		responseData = []string{"Placeholder for S3 data coming from march-summary.json"}
	}

	context.JSON(http.StatusOK, gin.H{
		"device_id": deviceID,
		"period":    period,
		"source":    map[bool]string{true: "DynamoDB", false: "S3"}[isHotTier(period)], 
		"data":      responseData,
	})
}

func isHotTier(period string) bool {
	switch period {
	case "1h", "24h", "5d", "7d":
		return true
	default:	
		return false  //it was deleted from dynamobd and now stored in S3
	}
}
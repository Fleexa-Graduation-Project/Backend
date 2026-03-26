package handlers

import (
    "log/slog"
    "net/http"
    "time"
    "fmt"

    "github.com/Fleexa-Graduation-Project/Backend/internal/devices"
    "github.com/Fleexa-Graduation-Project/Backend/internal/telemetry"
    "github.com/Fleexa-Graduation-Project/Backend/models"
	"github.com/Fleexa-Graduation-Project/Backend/internal/alerts"
    "github.com/gin-gonic/gin"
)

type DeviceHandler struct {
    StateStore     *devices.StateStore
    TelemetryStore *telemetry.TelemetryStore
    AlertStore     *alerts.AlertStore
}


func addLightStatus(payload map[string]interface{}, operationalState string) {
    switch operationalState {
    case "BRIGHT":
        payload["light_status"] = "Bright"
    case "DARK":
        payload["light_status"] = "Dark"
    case "NORMAL":
        payload["light_status"] = "Normal"
    }
}

func addTempStats(response gin.H, data []models.Telemetry, metric, deviceID string, now int64) {
    stats, err := telemetry.CalculateTempState(data, metric, now)
    if err != nil {
        slog.Warn("failed to calculate temp stats", "device_id", deviceID, "metric", metric, "error", err)
    }
    response["min"] = stats.Min
    response["max"] = stats.Max
    response["average"] = stats.Average
}

// handling GET /devices
func (handler *DeviceHandler) GetDevices(context *gin.Context) {
	
    states, err := handler.StateStore.GetAllStates(context.Request.Context())
    if err != nil {
        context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch device states"})
        return
    }
    for i := range states {
		states[i].Status = devices.ConnectionStatus(states[i].LastSeenAt)
        if states[i].Type == "light-sensor" {
            addLightStatus(states[i].Payload, states[i].OperationalState)
        }
    }
    context.JSON(http.StatusOK, gin.H{"data": states})
}

// showing last 5 Recent Events with its time - the Last Activity time - warning and alerts based on unlock time
func showDoorStats(payload map[string]interface{}, history []models.Telemetry, now int64) {
	if len(history) == 0 {
		payload["recent_events"] = []map[string]interface{}{}
		payload["last_activity_time"] = "No activity"
		payload["security_alert"] = "SAFE"
		return
	}

	// Format recent events for the UI
	payload["recent_events"] = telemetry.FormatDoorEvents(history)
	
	// Format last activity time (e.g., "10 mins ago")
	payload["last_activity_time"] = telemetry.TimeAgo(history[0].Timestamp, now)
	
	if lockState, ok := payload["lock_state"].(string); ok && lockState == "UNLOCKED" {
		minutesUnlocked := float64(now-history[0].Timestamp) / 60.0
		
		alertStatus := "SAFE"
		if minutesUnlocked > 15 {
			alertStatus = "CRITICAL_ALERT"
		} else if minutesUnlocked > 7 && minutesUnlocked <= 15 {
			alertStatus = "WARNING"
		}
		payload["security_alert"] = alertStatus
	} else {
		payload["security_alert"] = "SAFE"
	}
}

// getting normal state in door insights
func addDoorInsights(response gin.H, data []models.Telemetry, state *models.DeviceState, now int64) {
	avgUnlock := telemetry.CalculateAvgUnlock(data, now)
	response["average_unlock_minutes"] = avgUnlock

	
	normalDuration := 15.0 
	if userPref, ok := state.Payload["normal_unlock_duration"].(float64); 
    ok {
		normalDuration = userPref
	}
	
	if avgUnlock > normalDuration {
		response["unlock_duration_status"] = "Above Normal"
	} else {
		response["unlock_duration_status"] = "Normal"
	}
}

// handling GET /devices/:id
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

    state.Status = devices.ConnectionStatus(state.LastSeenAt)
    if state.Type == "light-sensor" {
        addLightStatus(state.Payload, state.OperationalState)
    }
    if state.Type == "door-actuator" {
		now := time.Now().Unix()
		//get the 5 most recent events
		recentHistory, dbErr := handler.TelemetryStore.GetTelemetryHistory(context.Request.Context(), deviceID, 5, 0)
		if dbErr != nil {
			slog.Warn("failed to fetch recent door history", "device_id", deviceID, "error", dbErr)
		}
		showDoorStats(state.Payload, recentHistory, now)
	}

    context.JSON(http.StatusOK, state)
}

// handling GET /devices/:id/telemetry?period=...&metric=...
func (handler *DeviceHandler) GetDeviceTelemetry(context *gin.Context) {
    deviceID := context.Param("id")
    period := context.DefaultQuery("period", "24h")
    metric := context.DefaultQuery("metric", "temp")

    now := time.Now().Unix()
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
        // Pass the period cutoff to DynamoDB 
        cutoff := telemetry.PeriodCutoff(now, period)
        rawData, dbErr := handler.TelemetryStore.GetTelemetryHistory(context.Request.Context(), deviceID, 0, cutoff)
        if dbErr != nil {
            context.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch telemetry history: device_id=%s, period=%s, error=%v", deviceID, period, dbErr),
            })
            return
        }

        response["source"] = "DynamoDB"
        response["data"] = telemetry.FilterTime(rawData, metric, period, now)

        if state.Type == "temp-sensor" {
            addTempStats(response, rawData, metric, deviceID, now)
        }

        if state.Type == "door-actuator" {
			addDoorInsights(response, rawData, state, now)
		}

    } else {
        response["source"] = "S3"
        response["data"] = []telemetry.ChartPoint{} // will handle s3 later

        if state.Type == "temp-sensor" {
            cutoff24h := now - 86400
            recentData, fetchErr := handler.TelemetryStore.GetTelemetryHistory(context.Request.Context(), deviceID, 0, cutoff24h)
            if fetchErr != nil {
                context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent history"})
                return
            }
            addTempStats(response, recentData, metric, deviceID, now)
        }
   
}

    context.JSON(http.StatusOK, response)
}

func (handler *DeviceHandler) GetDeviceAlerts(context *gin.Context) {
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

    alertList, err := handler.AlertStore.GetAlertsByDevice(context.Request.Context(), deviceID, 0)
    if err != nil {
        context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch alerts"})
        return
    }
    context.JSON(http.StatusOK, gin.H{"data": alertList})
}
func isHotTier(period string) bool {
    switch period {
    case "1h", "24h", "7d":
        return true
    default:
        return false
    }
}
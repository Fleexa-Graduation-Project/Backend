package telemetry

import (
	"math"
	"time"

	"github.com/Fleexa-Graduation-Project/Backend/models"
)

// DailyUsage represents a single bar on the UI Bar Chart
type DailyUsage struct {
	Date       string  `json:"date"`        // e.g., "2026-03-01"
	UsageHours float64 `json:"usage_hours"` // e.g., 4.5
}

// GroupACUsageByDay takes raw DynamoDB timestamps and formats them for the UI Bar Chart
func GroupACUsageByDay(history []models.Telemetry) []DailyUsage {
	dailyMap := make(map[string]float64)

	for _, record := range history {
		// Convert Unix timestamp to YYYY-MM-DD
		recordTime := time.Unix(record.Timestamp, 0)
		dayString := recordTime.Format("2006-01-02")

		// Check if power is ON. (Assuming each ping represents a 5-minute interval = 0.083 hours)
		if powerStr, exists := record.Payload["power"].(string); exists && powerStr == "ON" {
			dailyMap[dayString] += 0.083
		}
	}

	var result []DailyUsage
	for date, hours := range dailyMap {
		result = append(result, DailyUsage{
			Date:       date,
			UsageHours: math.Round(hours*10) / 10, 
		})
	}

	return result
}
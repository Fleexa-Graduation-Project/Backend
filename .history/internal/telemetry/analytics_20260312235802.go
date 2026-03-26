package telemetry

import (
	"cmp"
	"math"
	"slices"
	"time"

	"github.com/Fleexa-Graduation-Project/Backend/models"
)

type ChartPoint struct {
	Label string  `json:"label"` // x-axis
	Value float64 `json:"value"` // y-axis
}

// temp min max avg state
type TempState struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Average float64 `json:"average"`
}

// AggregateTimeSeries formats raw data into the ordered array needed by Flutter
func AggregateTimeSeries(history []models.Telemetry, metric string, period string) []ChartPoint {
	var timeFormat string
	switch period {
	case "24h":
		timeFormat = "15:00" // Group by hour
	case "7d":
		timeFormat = "2006-01-02" // Group by day
	case "1y":
		timeFormat = "2006-01" // Group by month
	default:
		timeFormat = "2006-01-02"
	}

	groupedData := make(map[string]float64)
	countMap := make(map[string]int)

	for _, record := range history {
		if val, exists := record.Payload[metric]; exists {
			recordTime := time.Unix(record.Timestamp, 0)
			timeLabel := recordTime.Format(timeFormat)

			// 1. Actuator Logic (e.g., Power ON -> total usage hours for the bar chart)
			if strVal, ok := val.(string); ok && strVal == "ON" {
				groupedData[timeLabel] += 0.083 // Assuming 5-minute pings
			}

			// 2. Sensor Logic (e.g., temp readings -> sum them up to average later for the line chart)
			if floatVal, ok := val.(float64); ok {
				groupedData[timeLabel] += floatVal
				countMap[timeLabel]++
			} else if intVal, ok := val.(int); ok {
				groupedData[timeLabel] += float64(intVal)
				countMap[timeLabel]++
			}
		}
	}

	// Convert the map into the final array
	var chartResult []ChartPoint
	for label, total := range groupedData {
		finalValue := total
		
		// If it's a sensor (we counted readings), divide the total by the count to get the average
		if count, wasSensor := countMap[label]; wasSensor && count > 0 {
			finalValue = total / float64(count)
		}

		chartResult = append(chartResult, ChartPoint{
			Label: label,
			Value: math.Round(finalValue*10) / 10, // Round to 1 decimal place
		})
	}
	
	slices.SortFunc(chartResult, func(a, b ChartPoint) int {
		return cmp.Compare(a.Label, b.Label)
	})

	return chartResult
}
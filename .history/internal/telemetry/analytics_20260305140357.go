package telemetry

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/Fleexa-Graduation-Project/Backend/models"
)

type ChartPoint struct {
	Label string  `json:"label"` // x coordinate (e.g., "10:00", "Mon")
	Value float64 `json:"value"` // y coordinate (e.g., 29.5)
}

// BasicStats holds the exact three numbers needed by specific UI screens (like the Temperature Sensor)
type BasicStats struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Average float64 `json:"average"`
}

// CalculateStats extracts the absolute min, max, and average from a set of raw telemetry
func CalculateStats(history []models.Telemetry, metric string) (BasicStats, error) {
	if len(history) == 0 {
		return BasicStats{}, fmt.Errorf("no data available to calculate stats")
	}

	overallMin := math.MaxFloat64
	overallMax := -math.MaxFloat64
	overallSum := 0.0
	overallCount := 0

	for _, record := range history {
		if val, exists := record.Payload[metric]; exists {
			var num float64
			if floatVal, ok := val.(float64); ok {
				num = floatVal
			} else if intVal, ok := val.(int); ok {
				num = float64(intVal)
			} else {
				continue // It's not a numeric sensor reading, skip math
			}

			if num < overallMin {
				overallMin = num
			}
			if num > overallMax {
				overallMax = num
			}
			overallSum += num
			overallCount++
		}
	}

	if overallCount == 0 {
		return BasicStats{}, fmt.Errorf("no valid numerical data found for metric: %s", metric)
	}

	return BasicStats{
		Min:     math.Round(overallMin*10) / 10,
		Max:     math.Round(overallMax*10) / 10,
		Average: math.Round((overallSum/float64(overallCount))*10) / 10,
	}, nil
}

// AggregateTimeSeries formats raw data into the X/Y array needed by the UI Insights charts
func AggregateTimeSeries(history []models.Telemetry, metric string, period string) []ChartPoint {
	var timeFormat string
	switch period {
	case "24h":
		timeFormat = "15:00"
	case "5d", "7d", "1m":
		timeFormat = "2006-01-02"
	case "1y":
		timeFormat = "2006-01"
	default:
		timeFormat = "2006-01-02"
	}

	groupedData := make(map[string]float64)
	countMap := make(map[string]int)

	for _, record := range history {
		if val, exists := record.Payload[metric]; exists {
			recordTime := time.Unix(record.Timestamp, 0)
			timeLabel := recordTime.Format(timeFormat)

			// 1. Actuator Logic (e.g., Power ON -> usage hours)
			if strVal, ok := val.(string); ok && strVal == "ON" {
				groupedData[timeLabel] += 0.083 // Assuming 5-minute intervals for prototype
			}

			// 2. Sensor Logic (e.g., numeric readings -> sums and counts for averaging)
			if floatVal, ok := val.(float64); ok {
				groupedData[timeLabel] += floatVal
				countMap[timeLabel]++
			} else if intVal, ok := val.(int); ok {
				groupedData[timeLabel] += float64(intVal)
				countMap[timeLabel]++
			}
		}
	}

	// converting mapped chart data into ordered array
	var chartResult []ChartPoint
	for label, total := range groupedData {
		finalValue := total
		// If it was a sensor, calculate the average for that specific time slice
		if count, wasSensor := countMap[label]; wasSensor && count > 0 {
			finalValue = total / float64(count)
		}

		chartResult = append(chartResult, ChartPoint{
			Label: label,
			Value: math.Round(finalValue*10) / 10,
		})
	}

	// sorting chronologically using modern go slices package
	slices.SortFunc(chartResult, func(a, b ChartPoint) int {
		return cmp.Compare(a.Label, b.Label)
	})

	return chartResult
}
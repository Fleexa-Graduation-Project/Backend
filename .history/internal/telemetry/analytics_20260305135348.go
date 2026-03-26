package telemetry

import (
	"cmp"
	"math"
	"slices"
	"time"

	"github.com/Fleexa-Graduation-Project/Backend/models"
)

type ChartPoint struct {
	Label string  `json:"label"` // x coordinate
	Value float64 `json:"value"` // y coordinate
}

type Analytics struct {
	Min        float64      `json:"min"`
	Max        float64      `json:"max"`
	Average    float64      `json:"average"`
	ChartData  []ChartPoint `json:"chart_data"`
}

// GenerateAnalytics processes raw telemetry into the exact unified payload the UI needs
func GenerateAnalytics(history []models.Telemetry, metric string, period string) UnifiedAnalytics {
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

	// Variables for overall stats
	overallMin := math.MaxFloat64
	overallMax := -math.MaxFloat64
	overallSum := 0.0
	overallCount := 0

	for _, record := range history {
		if val, exists := record.Payload[metric]; exists {
			recordTime := time.Unix(record.Timestamp, 0)
			timeLabel := recordTime.Format(timeFormat)

			// Safely extract the number
			var num float64
			if floatVal, ok := val.(float64); ok {
				num = floatVal
			} else if intVal, ok := val.(int); ok {
				num = float64(intVal)
			} else {
				continue // It's not a numeric sensor reading, skip math
			}

			// 1. Calculate overall Min/Max/Avg
			if num < overallMin {
				overallMin = num
			}
			if num > overallMax {
				overallMax = num
			}
			overallSum += num
			overallCount++

			// 2. Aggregate for the Chart
			groupedData[timeLabel] += num
			countMap[timeLabel]++
		}
	}

	// Handle case where no valid data was found
	if overallCount == 0 {
		return UnifiedAnalytics{Min: 0, Max: 0, Average: 0, ChartData: []ChartPoint{}}
	}

	// 3. Convert mapped chart data into ordered array
	var chartResult []ChartPoint
	for label, total := range groupedData {
		chartResult = append(chartResult, ChartPoint{
			Label: label,
			Value: math.Round((total/float64(countMap[label]))*10) / 10, // Average for that specific time slice, rounded
		})
	}

	// Sort chronologically (Go 1.26 feature)
	slices.SortFunc(chartResult, func(a, b ChartPoint) int {
		return cmp.Compare(a.Label, b.Label)
	})

	// Return the unified package!
	return UnifiedAnalytics{
		Min:       math.Round(overallMin*10) / 10,
		Max:       math.Round(overallMax*10) / 10,
		Average:   math.Round((overallSum/float64(overallCount))*10) / 10,
		ChartData: chartResult,
	}
}
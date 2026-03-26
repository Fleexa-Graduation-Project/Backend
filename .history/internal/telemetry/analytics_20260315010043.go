package telemetry

import (
	"fmt"
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


// Export PeriodCutoff so callers can pass the same cutoff to the DB query.
func PeriodCutoff(period string) int64 {
    switch period {
    case "24h":
        return time.Now().Add(-24 * time.Hour).Unix()
    case "7d":
        return time.Now().Add(-7 * 24 * time.Hour).Unix()
    case "1m":
        return time.Now().Add(-30 * 24 * time.Hour).Unix()
    default:
        return 0
    }
}

func FilterTime(history []models.Telemetry, metric string, period string) []ChartPoint {
    var timeFormat string
    switch period {
    case "24h":
        timeFormat = "15:04"
    case "7d", "1m":
        timeFormat = "2006-01-02"
    default:
        timeFormat = "2006-01-02"
    }

    cutoff := PeriodCutoff(period)

    // Pre-allocate maps with known upper bound to avoid internal rehashing
    var mapCapacity int
    switch period {
    case "24h":
        mapCapacity = 24
    case "7d":
        mapCapacity = 7
    case "1m":
        mapCapacity = 30
    default:
        mapCapacity = 30
    }

    groupedData := make(map[string]float64, mapCapacity)
    countMap := make(map[string]int, mapCapacity)

    for _, record := range history {
        if cutoff > 0 && record.Timestamp < cutoff {
            break 
        }

        if val, exists := record.Payload[metric]; exists {
            recordTime := time.Unix(record.Timestamp, 0)
            timeLabel := recordTime.Format(timeFormat)

            if strVal, ok := val.(string); ok && strVal == "ON" {
                groupedData[timeLabel] += 0.083
            }

            if floatVal, ok := val.(float64); ok {
                groupedData[timeLabel] += floatVal
                countMap[timeLabel]++
            } else if intVal, ok := val.(int); ok {
                groupedData[timeLabel] += float64(intVal)
                countMap[timeLabel]++
            }
        }
    }

    chartResult := make([]ChartPoint, 0, len(groupedData))
    for label, total := range groupedData {
        finalValue := total
        if count, wasSensor := countMap[label]; wasSensor && count > 0 {
            finalValue = total / float64(count)
        }
        chartResult = append(chartResult, ChartPoint{
            Label: label,
            Value: math.Round(finalValue*10) / 10,
        })
    }

    slices.SortFunc(chartResult, func(a, b ChartPoint) int {
        return cmp.Compare(a.Label, b.Label)
    })

    return chartResult
}

func CalculateTempState(history []models.Telemetry, metric string) (TempState, error) {
    if len(history) == 0 {
        return TempState{}, fmt.Errorf("no data")
    }

    overallMin := math.MaxFloat64
    overallMax := -math.MaxFloat64
    overallSum := 0.0
    overallCount := 0
    cutoffTime := time.Now().Add(-24 * time.Hour).Unix()

    for _, record := range history {
        if record.Timestamp < cutoffTime {
            break 
        }

        if val, exists := record.Payload[metric]; exists {
            var num float64
            if floatVal, ok := val.(float64); ok {
                num = floatVal
            } else if intVal, ok := val.(int); ok {
                num = float64(intVal)
            } else {
                continue
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
        return TempState{Min: 0, Max: 0, Average: 0}, nil
    }

    return TempState{
        Min:     math.Round(overallMin*10) / 10,
        Max:     math.Round(overallMax*10) / 10,
        Average: math.Round((overallSum/float64(overallCount))*10) / 10,
    }, nil
}
package devices

var Rules = map[string]DeviceRules{
	"temp-sensor": {
		ExtractOperational: func(payload map[string]interface{}) string {
			if temp, ok := payload["temp"].(float64); 
			ok {
				if temp > 30 {
					return "HOT"
				}
				if temp < 18 {
					return "COLD"
				}
				return "NORMAL"
			}
			return "UNKNOWN"
		},
		EvaluateHealth: func(op string) string {
			switch op {
			case "HOT":
				return "DEGRADED"
			case "COLD", "NORMAL":
				return "HEALTHY"
			default:
				return "DEGRADED"
			}
		},
	},

	"light-sensor": {
		ExtractOperational: func(payload map[string]interface{}) string {
			if level, ok := payload["light_level"].(float64);
			ok {
				if level > 600 {
					return "BRIGHT"
				}
				if level < 200 {
					return "DIM"
				}
				return "NORMAL"
			}
			return "UNKNOWN"
		},
		EvaluateHealth: func(op string) string {
			return "HEALTHY"
		},
	},

	"door-actuator": {
		ExtractOperational: func(payload map[string]interface{}) string {
			if state, ok := payload["lock_state"].(string); ok {
				return state // LOCKED or UNLOCKED
			}
			return "UNKNOWN"
		},
		EvaluateHealth: func(op string) string {
			return "HEALTHY"
		},
	},
	"door-sensor": {
    ExtractOperational: func(payload map[string]interface{}) string {
        if open, ok := payload["open"].(bool); 
		ok {
            if open {
                return "OPEN"
            }
            return "CLOSED"
        }
		if openStr, ok := payload["open"].(string); ok {
				// Handle "true", "TRUE", "True", "open"
				lower := strings.ToLower(openStr)
				if lower == "true" || lower == "open" {
					return "OPEN"
				}
				return "CLOSED"
			}
        return "UNKNOWN"
    },
    EvaluateHealth: func(op string) string {
		if op == "UNKNOWN" { 
			return "DEGRADED"
		} 
        return "HEALTHY"
    },
},

}

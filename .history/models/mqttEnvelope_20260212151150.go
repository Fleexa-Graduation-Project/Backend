package models

// incoming message structure
type MQTTEnvelope struct {
	DeviceID  string                 `json:"device_id"`
	Timestamp int64                  `json:"timestamp"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
}

// TelemetryReading is a single reading inside a batch packet
type TelemetryReading struct {
	Timestamp int64                  `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// MQTTBatchEnvelope wraps a batch of telemetry readings from one device.
// Devices send these "packets" to reduce per-message overhead.
type MQTTBatchEnvelope struct {
	DeviceID string             `json:"device_id"`
	Timestamp int64             `json:"timestamp"`
	Type     string             `json:"type"`
	Readings []TelemetryReading `json:"readings"`
}
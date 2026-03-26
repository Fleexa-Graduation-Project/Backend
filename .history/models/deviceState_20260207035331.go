package models

type DeviceState struct {
	DeviceID     string `dynamodbav:"device_id"`
	LastSeen   int64  `dynamodbav:"last_seen_at"`
	Status       string `dynamodbav:"status"`   // online-offlice OFFLINE
	Health       string `dynamodbav:"health"`   // HEALTHY | DEGRADED | CRITICAL
	UpdatedAt    int64  `dynamodbav:"updated_at"`
}

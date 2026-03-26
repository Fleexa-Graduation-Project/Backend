package models

type DeviceState struct {
	DeviceID     string `dynamodbav:"device_id"`
	LastSeenAt   int64  `dynamodbav:"last_seen_at"`
	Status       string `dynamodbav:"status"`   // online - offline 
	Health       string `dynamodbav:"health"`   // healthy - DEGRADED - critical
	LastUpdated int64  `dynamodbav:"updated_at"`
}

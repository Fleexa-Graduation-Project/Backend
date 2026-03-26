package models

type DeviceState struct {
	DeviceID     string `dynamodbav:"device_id"`
	Type            string `dynamodbav:"type"`
	
	Status       string `dynamodbav:"status"`   // online - offline 
	Health       string `dynamodbav:"health"`   // healthy - degraded - critical
	LastSeenAt   int64  `dynamodbav:"last_seen_at"`
	LastUpdated int64  `dynamodbav:"updated_at"`
}

package models

type DeviceState struct {
	
	
	OperationalState string `dynamodbav:"operational_state"` // based on device: LOCKED-HOT-BRIGHT-OFF etc.

	Health       string `dynamodbav:"health"`   // healthy - degraded - critical
}
package models

type DeviceState struct {
	DeviceID     string                 `json:"device_id" dynamodbav:"device_id"`
	Type         string                 `json:"type" dynamodbav:"type"`
	Status       string                 `json:"status" dynamodbav:"status"` // ONLINE - OFFLINE
	CurrentState map[string]interface{} `json:"current_state" dynamodbav:"current_state"` // based on device: LOCKED-HOT-BRIGHT-OFF etc.
	LastSeenAt   int64                  `json:"last_seen_at" dynamodbav:"last_seen_at"`
	LastUpdated  int64                  `json:"-" dynamodbav:"updated_at"` 
}
package models

type DeviceState struct {
	DeviceID     string `dynamodbav:"device_id"`
	Type            string `dynamodbav:"type"`
	
	Status       string `dynamodbav:"status"`   // online - offline 
	OperationalState string `dynamodbav:"operational_state"` // based on device: LOCKED-HOT-BRIGHT-OFF etc.

	Health       string `dynamodbav:"health"`   // healthy - degraded - critical
	LastSeenAt   int64  `dynamodbav:"last_seen_at"`
	LastUpdated int64  `dynamodbav:"updated_at"`
}
package models

type DeviceState struct {
	DeviceID         string                 `json:"device_id" dynamodbav:"device_id"`
	Type             string                 `json:"type" dynamodbav:"type"`
	Status           string                 `json:"status" dynamodbav:"status"` // ONLINE or OFFLINE
	OperationalState string                 `json:"operational_state" dynamodbav:"operational_state"`
	Health           string                 `json:"health" dynamodbav:"health"`
	Payload          map[string]interface{} `json:"payload" dynamodbav:"payload"` // Raw sensor data (temp, gas_level)
	LastSeenAt       int64                  `json:"last_seen_at" dynamodbav:"last_seen_at"`
	LastUpdated      int64                  `json:"-" dynamodbav:"updated_at"` // Hidden from API
}
package validation

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/models"
)

const MaxBatchSize = 500 // hard cap on readings per batch packet

// IsBatchTelemetry returns true when the event payload carries a "readings"
// array instead of a single "payload" object. This lets the handler pick
// the batch fast-path before any heavy validation runs.
func IsBatchTelemetry(event map[string]interface{}) bool {
	payload, ok := event["payload"]
	if !ok {
		return false
	}

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return false
	}

	readings, ok := payloadMap["readings"]
	if !ok {
		return false
	}

	_, ok = readings.([]interface{})
	return ok
}

// ValidateBatchMessage validates a batch telemetry event and returns the
// device ID together with the decoded batch envelope.
func ValidateBatchMessage(event map[string]interface{}) (
	deviceID string,
	batchEnvelope models.MQTTBatchEnvelope,
	err error,
) {
	// 1. raw event structure
	topic, payloadRaw, err := validateEvent(event)
	if err != nil {
		return "", batchEnvelope, err
	}

	// 2. topic must point to telemetry
	var messageType string
	deviceID, messageType, err = validateTopic(topic)
	if err != nil {
		return "", batchEnvelope, err
	}

	if messageType != "telemetry" {
		return "", batchEnvelope,
			fmt.Errorf("%w: batch ingestion only supported for telemetry", ErrInvalidEvent)
	}

	// 3. decode into batch envelope
	if err := decodeBatchEnvelope(payloadRaw, &batchEnvelope); err != nil {
		return "", batchEnvelope, err
	}

	// 4. validate envelope-level fields
	if err := validateBatchEnvelope(batchEnvelope, deviceID); err != nil {
		return "", batchEnvelope, err
	}

	// 5. validate every reading
	if err := validateReadings(batchEnvelope.Type, batchEnvelope.Readings); err != nil {
		return "", batchEnvelope, err
	}

	return deviceID, batchEnvelope, nil
}

// ── helpers ──────────────────────────────────────────────────────────────

func decodeBatchEnvelope(payloadRaw interface{}, env *models.MQTTBatchEnvelope) error {
	bytes, err := json.Marshal(payloadRaw)
	if err != nil {
		return fmt.Errorf("%w: batch payload marshal failed", ErrInvalidEnvelope)
	}

	// 512 KB ceiling for batch payloads (vs. 32 KB for single messages)
	if len(bytes) > 512*1024 {
		return fmt.Errorf("%w: batch payload too large (max 512 KB)", ErrInvalidPayload)
	}

	if err := json.Unmarshal(bytes, env); err != nil {
		return fmt.Errorf("%w: batch payload unmarshal failed", ErrInvalidEnvelope)
	}

	return nil
}

func validateBatchEnvelope(env models.MQTTBatchEnvelope, topicDeviceID string) error {
	if env.DeviceID == "" {
		return fmt.Errorf("%w: missing device_id", ErrInvalidEnvelope)
	}

	if env.DeviceID != topicDeviceID {
		return fmt.Errorf("%w: device_id mismatch", ErrInvalidEnvelope)
	}

	if env.Type == "" {
		return fmt.Errorf("%w: missing device type", ErrInvalidEnvelope)
	}

	if len(env.Readings) == 0 {
		return fmt.Errorf("%w: empty readings array", ErrInvalidEnvelope)
	}

	if len(env.Readings) > MaxBatchSize {
		return fmt.Errorf("%w: batch exceeds max size of %d readings",
			ErrInvalidPayload, MaxBatchSize)
	}

	return nil
}

func validateReadings(deviceType string, readings []models.TelemetryReading) error {
	rules, ok := devices.Rules[deviceType]
	if !ok {
		return fmt.Errorf("%w: unknown device type", ErrInvalidPayload)
	}

	now := time.Now().Unix()

	for i, r := range readings {
		if r.Timestamp == 0 {
			return fmt.Errorf("%w: reading[%d] missing timestamp", ErrInvalidPayload, i)
		}

		if r.Timestamp > now+60 {
			return fmt.Errorf("%w: reading[%d] timestamp in the future", ErrInvalidPayload, i)
		}

		if len(r.Payload) == 0 {
			return fmt.Errorf("%w: reading[%d] empty payload", ErrInvalidPayload, i)
		}

		op := rules.ExtractOperational(r.Payload)
		if op == "UNKNOWN" {
			return fmt.Errorf("%w: reading[%d] payload does not match device type",
				ErrInvalidPayload, i)
		}

		// type-specific validation
		switch deviceType {
		case "temp-sensor":
			if _, ok := r.Payload["temp"].(float64); !ok {
				return fmt.Errorf("%w: reading[%d] temp must be numeric",
					ErrInvalidPayload, i)
			}
		}
	}

	return nil
}

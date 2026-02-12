package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/models"
)
var (
	InvalidEvent    = errors.New("invalid event")
	InvalidTopic    = errors.New("invalid topic")
	InvalidEnvelope = errors.New("invalid envelope")
	InvalidPayload  = errors.New("invalid payload")
)
// validating incoming MQTT messages
func ValidateMessage(event map[string]interface{},) (
	deviceID string,
	messageType string,
	envelope models.MQTTEnvelope,
	err error,
){

	//validating raw event
	topic, payloadRaw, err := validateEvent(event)
	if err != nil {
		return "", "", envelope, err
	}

	// validating topic
	deviceID, messageType, err = validateTopic(topic)
	if err != nil {
		return "", "", envelope, err
	}

	// decoding payload into envelope
	if err := decodeEnvelope(payloadRaw, &envelope); err != nil {
		return "", "", envelope, err
	}

	// validating envelope fields
	if err := validateEnvelope(envelope, deviceID); err != nil {
		return "", "", envelope, err
	}

	// validating payload structure based on message type
	if err := validatePayload(envelope.Type, envelope.Payload); err != nil {
		return "", "", envelope, err
	}

	return deviceID, messageType, envelope, nil
}


func validateEvent(event map[string]interface{}) (string, interface{}, error) {
	topic, ok := event["topic"].(string)
	if !ok || topic == "" {
		return "", nil, fmt.Errorf("%w: missing or invalid topic", InvalidEvent)
	}

	payload, ok := event["payload"]
	if !ok {
		return "", nil, fmt.Errorf("%w: missing payload", InvalidEvent)
	}

	return topic, payload, nil
}

func validateTopic(topic string) (string, string, error) {
	parts := strings.Split(topic, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("%w: expected devices/{id}/{type}", InvalidTopic)
	}

	if parts[0] != "devices" {
		return "", "", fmt.Errorf("%w: invalid topic root", InvalidTopic)
	}

	deviceID := parts[1]
	messageType := parts[2]

	if deviceID == "" {
		return "", "", fmt.Errorf("%w: empty device id", InvalidTopic)
	}

	switch messageType {
	case "telemetry", "alerts":
		return deviceID, messageType, nil
	default:
		return "", "", fmt.Errorf("%w: unsupported message type", InvalidTopic)
	}
}


func decodeEnvelope(payloadRaw interface{}, env *models.MQTTEnvelope) error {
	bytes, err := json.Marshal(payloadRaw)
	if err != nil {
		return fmt.Errorf("%w: payload marshal failed", InvalidEnvelope)
	}

	if err := json.Unmarshal(bytes, env); err != nil {
		return fmt.Errorf("%w: payload unmarshal failed", InvalidEnvelope)
	}

	return nil
}

func validateEnvelope(env models.MQTTEnvelope, topicDeviceID string) error {
	if env.DeviceID == "" {
		return fmt.Errorf("%w: missing device_id", InvalidEnvelope)
	}

	if env.DeviceID != topicDeviceID {
		return fmt.Errorf("%w: device_id mismatch", InvalidEnvelope)
	}

	if env.Timestamp <= 0 {
		return fmt.Errorf("%w: invalid timestamp", InvalidEnvelope)
	}

	if env.Type == "" {
		return fmt.Errorf("%w: missing device type", InvalidEnvelope)
	}

	if len(env.Payload) == 0 {
		return fmt.Errorf("%w: empty payload", InvalidEnvelope)
	}

	return nil
}

func validatePayload(deviceType string, payload map[string]interface{}) error {
	rules, ok := devices.Rules[deviceType]
	if !ok {
		return fmt.Errorf("%w: unknown device type", InvalidPayload)
	}

	op := rules.ExtractOperational(payload)
	if op == "UNKNOWN" {
		return fmt.Errorf("%w: payload does not match device type", ErrInvalidPayload)
	}

	return nil
}




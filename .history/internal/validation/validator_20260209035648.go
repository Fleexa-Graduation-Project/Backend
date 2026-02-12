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
		return "", nil, fmt.Errorf("%w: missing or invalid topic", ErrInvalidEvent)
	}

	payload, ok := event["payload"]
	if !ok {
		return "", nil, fmt.Errorf("%w: missing payload", ErrInvalidEvent)
	}

	return topic, payload, nil
}




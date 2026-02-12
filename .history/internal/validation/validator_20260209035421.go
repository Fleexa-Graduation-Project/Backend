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
// Validating incoming MQTT messages
func ValidateMessage(event map[string]interface{},) (
	deviceID string,
	messageType string,
	envelope models.MQTTEnvelope,
	err error,
){

	//Validating raw event
	topic, payloadRaw, err := validateEvent(event)
	if err != nil {
		return "", "", envelope, err
	}

	// 2️⃣ Validate topic
	deviceID, messageType, err = validateTopic(topic)
	if err != nil {
		return "", "", envelope, err
	}

	// 3️⃣ Decode payload into envelope
	if err := decodeEnvelope(payloadRaw, &envelope); err != nil {
		return "", "", envelope, err
	}

	// 4️⃣ Validate envelope consistency
	if err := validateEnvelope(envelope, deviceID); err != nil {
		return "", "", envelope, err
	}

	// 5️⃣ Validate payload semantics
	if err := validatePayload(envelope.Type, envelope.Payload); err != nil {
		return "", "", envelope, err
	}

	return deviceID, messageType, envelope, nil
}



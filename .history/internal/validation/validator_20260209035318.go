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
func ValidateMessage(event map[string]interface{},
) (
	deviceID string,
	messageType string,
	envelope models.MQTTEnvelope,
	err error,
)



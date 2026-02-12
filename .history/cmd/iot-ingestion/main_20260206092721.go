package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/Fleexa-Graduation-Project/Backend/models"
	"github.com/Fleexa-Graduation-Project/Backend/internal/telemetry"
	"github.com/Fleexa-Graduation-Project/Backend/internal/alerts"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/db"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/logger")

var (
	log *slog.Logger
	telemetryStore *telemetry.TelemetryStore
	alertStore    *alerts.AlertStore
)

// for Cold Start
func init() {

	log = logger.InitLogger()
	log.Info("IoT Ingestion: Cold Start")

	if err := db.Init(context.Background()); err != nil {
		log.Error("Failed to connect to DynamoDB", "error", err)
		panic(fmt.Sprintf("DB Init failed: %v", err))
	}
}

func parseTopic(topic string) (deviceID string, messageType string, err error) {
	topicParts := strings.Split(topic, "/")
	if len(topicParts) != 3 {
		return "", "", fmt.Errorf("invalid topic format: %s", topic)
	}

	if topicParts[0] != "devices" {
		return "", "", fmt.Errorf("invalid topic start: %s", topic)
	}

	return topicParts[1], topicParts[2], nil
}



func HandleRequest(ctx context.Context, event events.IoTCoreMessage) error {
	log.Info("message received", "topic", event.Topic)

	// 1. Parse topic
	deviceIDFromTopic, messageType, err := parseTopic(event.Topic)
	if err != nil {
		log.Error("Topic parsing failed", "error", err)
		return err
	}

	// 2. Parse JSON payload
	var envelope MQTTEnvelope
	if err := json.Unmarshal(event.Payload, &envelope); err != nil {
		log.Error("Invalid JSON payload", "error", err)
		return err
	}

	// 3. Validate consistency
	if envelope.DeviceID != deviceIDFromTopic {
		return fmt.Errorf("device_id mismatch: topic=%s payload=%s",
			deviceIDFromTopic, envelope.DeviceID)
	}

	if envelope.Timestamp == 0 {
		return fmt.Errorf("missing timestamp")
	}

	if len(envelope.Payload) == 0 {
		return fmt.Errorf("empty payload")
	}

	// 4. Route message
	switch messageType {

	case "telemetry":
		store, err := telemetry.NewTelemetryStore()
		if err != nil {
			return err
		}

		data := models.Telemetry{
			DeviceID:  envelope.DeviceID,
			Timestamp: envelope.Timestamp,
			Type:      envelope.Type,
			Payload:   envelope.Payload,
		}

		return store.SaveTelemetry(ctx, data)

	case "alerts":
		store, err := alerts.NewAlertStore()
		if err != nil {
			return err
		}

		severity, ok := envelope.Payload["severity"].(string)
		if !ok {
			return fmt.Errorf("alert missing severity")
		}

		alert := models.Alert{
			DeviceID:  envelope.DeviceID,
			Timestamp: envelope.Timestamp,
			Type:      envelope.Type,
			Severity:  severity,
			Payload:   envelope.Payload,
		}

		return store.SaveAlert(ctx, alert)

	default:
		return fmt.Errorf("unknown message type: %s", messageType)
	}
}


func main() {

	lambda.Start(HandleRequest)
}
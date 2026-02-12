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
	"github.com/Fleexa-Graduation-Project/Backend/pkg/logger"

)

var (
	log *slog.Logger
	telemetryStore *telemetry.TelemetryStore
	alertStore    *alerts.AlertStore
)

// incoming message structure
type MQTTEnvelope struct {
	DeviceID  string                 `json:"device_id"`
	Timestamp int64                  `json:"timestamp"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
}


// for Cold Start
func init() {

	log = logger.InitLogger()
	log.Info("IoT Ingestion: Cold Start")

// Init DynamoDB 
	if err := db.Init(context.Background()); err != nil {
		log.Error("Failed to initialize DynamoDB", "error", err)
		panic(err)
	}

	var err error

	telemetryStore, err = telemetry.NewTelemetryStore()
	if err != nil {
		panic(fmt.Errorf("failed to init telemetry store: %w", err))
	}

	alertStore, err = alerts.NewAlertStore()
	if err != nil {
		panic(fmt.Errorf("failed to init alert store: %w", err))
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

	// topic Parsing 
	deviceID, messageType, err := parseTopic(event.Topic)
	if err != nil {
		log.Error("Topic parsing failed", "error", err)
		return err
	}

	//payload Parsing
	var  (envelope MQTTEnvelope
		) 
	if err := json.Unmarshal(event.Payload, &envelope); err != nil {
		log.Error("Invalid payload", "error", err)
		return err
	}

	// 3. Validating consistency
	if envelope.DeviceID != deviceID {
		return fmt.Errorf("No match for device_id: topic=%s payload=%s",
			deviceID, envelope.DeviceID)
	}

	if envelope.Timestamp == 0 {
		return fmt.Errorf("missing timestamp")
	}

	if len(envelope.Payload) == 0 {
		return fmt.Errorf("empty payload")
	}

	// message Routing
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
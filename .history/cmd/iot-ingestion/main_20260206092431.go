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
		return "", "", fmt.Errorf("invalid topic prefix: %s", topic)
	}

	return parts[1], parts[2], nil
}

func HandleRequest(ctx context.Context, event events.IoTButtonEvent) (string, error) {

	log.Info("Received new IoT message", "event_data", event)

	// TODO: will add the DynamoDB save logic here later

	return "Success", nil
}

func main() {

	lambda.Start(HandleRequest)
}
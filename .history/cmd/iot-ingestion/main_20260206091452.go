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
)

// for Cold Start
func init() {

	log = logger.InitLogger()
	log.Info("IoT Ingestion: Cold Start")
}

func HandleRequest(ctx context.Context, event events.IoTButtonEvent) (string, error) {

	log.Info("Received new IoT message", "event_data", event)

	// TODO: will add the DynamoDB save logic here later

	return "Success", nil
}

func main() {

	lambda.Start(HandleRequest)
}
package main

import (
	"context"
	"log/slog"

	"github.com/Fleexa-Graduation-Project/Backend
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	log *slog.Logger
)

// for Cold Start
func init() {

	log = logger.InitLogger()
	log.Info("IoT Ingestion Service: Cold Start Initialization")
}

func HandleRequest(ctx context.Context, event events.IoTButtonEvent) (string, error) {

	log.Info("Received new IoT message", "event_data", event)

	// TODO: will add the DynamoDB save logic here later

	return "Success", nil
}

func main() {

	lambda.Start(HandleRequest)
}
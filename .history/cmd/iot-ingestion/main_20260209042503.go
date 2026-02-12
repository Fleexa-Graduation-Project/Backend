package main

import (
	"context"
	"fmt"
	"log/slog"
	"errors"


	"github.com/aws/aws-lambda-go/lambda"
	"github.com/Fleexa-Graduation-Project/Backend/models"
	"github.com/Fleexa-Graduation-Project/Backend/internal/telemetry"
	"github.com/Fleexa-Graduation-Project/Backend/internal/alerts"
	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/db"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/logger"
	"github.com/Fleexa-Graduation-Project/Backend/internal/validation"

)

var (
	log *slog.Logger
	telemetryStore *telemetry.TelemetryStore
	alertStore    *alerts.AlertStore
	stateStore *devices.StateStore

)


// for Cold Start
func init() {

	log = logger.InitLogger()
	log.Info("IoT Ingestion: Cold Start")

// Init DynamoDB 
	if err := db.NewDynamoDBClient(context.Background()); err != nil {
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


	stateStore, err = devices.NewStateStore()
	if err != nil {
		panic(fmt.Errorf("failed to init device state store: %w", err))
	}


	

}



func HandleRequest(ctx context.Context, event map[string]interface{}) (err error) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("CRITICAL: Lambda Panic Recovered", "panic", r)
			err = fmt.Errorf("internal server error")
		}
	}()

	deviceID, messageType, envelope, err :=
		validation.ValidateMessage(event)
	if err != nil {
		if errors.Is(err, validation.ErrInvalidPayload) {
	log.Warn("Invalid payload", "error", err)
	return err
}

if errors.Is(err, validation.ErrInvalidTopic) {
	log.Warn("Invalid topic", "error", err)
	return err
}

log.Error("Unexpected validation error", "error", err)
		return err
	}

	// message Routing
	switch messageType {

	case "telemetry":

		data := models.Telemetry{
			DeviceID:  envelope.DeviceID,
			Timestamp: envelope.Timestamp,
			Type:      envelope.Type,
			Payload:   envelope.Payload,
		}

		log.Info("Saving Telemetry", "device_id", deviceID)
		if err := telemetryStore.SaveTelemetry(ctx, data); err != nil {
			log.Error("Failed to save telemetry", "error", err)
			return err
		}
		

		return stateStore.UpdateFromTelemetry(ctx, data)


	case "alerts":

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

		if err := alertStore.SaveAlert(ctx, alert); err != nil {
			return err
		}
		
		//updating Heartbeat so the device stays "ONLINE"
		return stateStore.UpdateHeartbeat(ctx, deviceID)
	default:
		return fmt.Errorf("unknown message type: %s", messageType)
	}
}


func main() {

	lambda.Start(HandleRequest)
}


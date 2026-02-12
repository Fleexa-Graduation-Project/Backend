package ingestion
import (
	"context"
	"fmt"
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/Fleexa-Graduation-Project/Backend/internal/validation"
	"github.com/Fleexa-Graduation-Project/Backend/internal/telemetry"
	
)

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

	switch {
	case errors.Is(err, validation.ErrInvalidEvent),
		errors.Is(err, validation.ErrInvalidEnvelope):
		log.Warn("Invalid message envelope", "error", err)
		return err

	case errors.Is(err, validation.ErrInvalidPayload):
		log.Warn("Invalid payload", "device_id", envelope.DeviceID, "error", err,)
		return err


	case errors.Is(err, validation.ErrInvalidTopic):
		log.Warn("Invalid topic", "error", err)
		return err

	default:
		log.Error("Unexpected validation error", "error", err)
		return err
	}
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

		severity, _ := envelope.Payload["severity"].(string)

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

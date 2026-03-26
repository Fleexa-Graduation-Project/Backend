package ingestion

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Fleexa-Graduation-Project/Backend/models"
	"github.com/Fleexa-Graduation-Project/Backend/internal/telemetry"
	"github.com/Fleexa-Graduation-Project/Backend/internal/alerts"
	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/internal/validation"
)
var (
	log *slog.Logger
	telemetryStore *telemetry.TelemetryStore
	alertStore *alerts.AlertStore
	stateStore *devices.StateStore
)
type Service struct {
	Logger         *slog.Logger
	TelemetryStore *telemetry.TelemetryStore
	AlertStore     *alerts.AlertStore
	StateStore     *devices.StateStore
}
func (s *Service) HandleRequest(ctx context.Context, event map[string]interface{}) (err error) {

	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error("CRITICAL: Lambda Panic Recovered", "panic", r)
			err = fmt.Errorf("internal server error")
		}
	}()

	deviceID, messageType, envelope, err :=
		validation.ValidateMessage(event)
	if err != nil {

		switch {
		case errors.Is(err, validation.ErrInvalidEvent),
			errors.Is(err, validation.ErrInvalidEnvelope):
			s.Logger.Warn("Invalid message envelope", "error", err)
			return err

		case errors.Is(err, validation.ErrInvalidPayload):
			s.Logger.Warn("Invalid payload", "device_id", envelope.DeviceID, "error", err)
			return err

		case errors.Is(err, validation.ErrInvalidTopic):
			s.Logger.Warn("Invalid topic", "error", err)
			return err

		default:
			s.Logger.Error("Unexpected validation error", "error", err)
			return err
		}
	}

	switch messageType {

	case "telemetry":

		data := models.Telemetry{
			DeviceID:  envelope.DeviceID,
			Timestamp: envelope.Timestamp,
			Type:      envelope.Type,
			Payload:   envelope.Payload,
		}

		s.Logger.Info("Saving Telemetry", "device_id", deviceID)

		if err := s.TelemetryStore.SaveTelemetry(ctx, data); err != nil {
			s.Logger.Error("Failed to save telemetry", "error", err)
			return err
		}

		return s.StateStore.UpdateFromTelemetry(ctx, data)

	case "alerts":

		severity, _ := envelope.Payload["severity"].(string)

		alert := models.Alert{
			DeviceID:  envelope.DeviceID,
			Timestamp: envelope.Timestamp,
			Type:      envelope.Type,
			Severity:  severity,
			Payload:   envelope.Payload,
		}

		if err := s.AlertStore.SaveAlert(ctx, alert); err != nil {
			return err
		}

		return s.StateStore.UpdateHeartbeat(ctx, deviceID)

	default:
		return fmt.Errorf("unknown message type: %s", messageType)
	}
}

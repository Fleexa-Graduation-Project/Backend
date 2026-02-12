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
		return s.handleTelemetry(ctx, deviceID, envelope)

	case "alerts":
		return s.handleAlert(ctx, deviceID, envelope)

	default:
		return fmt.Errorf("unknown message type: %s", messageType)
	}
}

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
func (s *Service) handleTelemetry(ctx context.Context, deviceID string, envelope *validation.Envelope) error {

	readingsRaw, isBatch := envelope.Payload["readings"]

	// ---- SINGLE READING ----
	if !isBatch {

		data := models.Telemetry{
			DeviceID:  envelope.DeviceID,
			Timestamp: envelope.Timestamp,
			Type:      envelope.Type,
			Payload:   envelope.Payload,
		}

		if err := s.TelemetryStore.SaveTelemetry(ctx, data); err != nil {
			s.Logger.Error("Failed to save telemetry", "error", err)
			return err
		}

		return s.StateStore.UpdateFromTelemetry(ctx, data)
	}

	// ---- BATCH MODE ----

	readingsSlice, ok := readingsRaw.([]interface{})
	if !ok {
		return fmt.Errorf("invalid readings format")
	}

	if len(readingsSlice) == 0 {
		return fmt.Errorf("empty telemetry batch")
	}

	var latestTelemetry models.Telemetry
	var savedCount int

	for _, r := range readingsSlice {

		rMap, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		timestampFloat, ok := rMap["timestamp"].(float64)
		if !ok {
			continue
		}

		payloadMap, ok := rMap["payload"].(map[string]interface{})
		if !ok {
			continue
		}

		telemetryData := models.Telemetry{
			DeviceID:  envelope.DeviceID,
			Timestamp: int64(timestampFloat),
			Type:      envelope.Type,
			Payload:   payloadMap,
		}

		if err := s.TelemetryStore.SaveTelemetry(ctx, telemetryData); err != nil {
			s.Logger.Error("Failed to save telemetry item", "error", err)
			continue
		}

		savedCount++

		// Track latest timestamp
		if telemetryData.Timestamp > latestTelemetry.Timestamp {
			latestTelemetry = telemetryData
		}
	}

	if savedCount == 0 {
		return fmt.Errorf("no valid telemetry readings saved")
	}

	// Update Device State ONLY ONCE using the latest reading
	return s.StateStore.UpdateFromTelemetry(ctx, latestTelemetry)
}
func (s *Service) handleAlert(ctx context.Context, deviceID string, envelope *validation.Envelope) error {

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
}

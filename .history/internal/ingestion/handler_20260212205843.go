package ingestion

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Fleexa-Graduation-Project/Backend/internal/alerts"
	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/internal/telemetry"
	"github.com/Fleexa-Graduation-Project/Backend/internal/validation"
	"github.com/Fleexa-Graduation-Project/Backend/models"
)

// Service holds dependencies for the ingestion logic
type Service struct {
	Logger         *slog.Logger
	TelemetryStore *telemetry.TelemetryStore
	AlertStore     *alerts.AlertStore
	StateStore     *devices.StateStore
}

func (s *Service) HandleRequest(ctx context.Context, event map[string]interface{}) (err error) {
	// Panic Recovery Shield
	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error("CRITICAL: Lambda Panic Recovered", "panic", r)
			err = fmt.Errorf("internal server error")
		}
	}()

	// 1. Validate the Message (Now returns isBatch!)
	deviceID, messageType, envelope, isBatch, err := validation.ValidateMessage(event)
	if err != nil {
		s.logValidationError(err, envelope.DeviceID)
		return err
	}

	// 2. Route based on Message Type
	switch messageType {
	case "telemetry":
		return s.handleTelemetry(ctx, deviceID, envelope, isBatch)

	case "alerts":
		return s.handleAlert(ctx, deviceID, envelope)

	default:
		return fmt.Errorf("unknown message type: %s", messageType)
	}
}

// handleTelemetry processes both Single and Batch telemetry messages
func (s *Service) handleTelemetry(ctx context.Context, deviceID string, envelope models.MQTTEnvelope, isBatch bool) error {
	
	// --- PATH A: BATCH PROCESSING ---
	if isBatch {
		items, ok := envelope.Payload["items"].([]interface{})
		if !ok {
			return fmt.Errorf("invalid batch format: items is not a list")
		}

		s.Logger.Info("Processing Batch Telemetry", "device_id", deviceID, "count", len(items))

		var telemetryList []models.Telemetry
		var latestReading models.Telemetry

		// Loop through the batch
		for _, itemRaw := range items {
			itemMap, ok := itemRaw.(map[string]interface{})
			if !ok {
				s.Logger.Warn("Skipping invalid item in batch", "device_id", deviceID)
				continue
			}

			// Validate individual item structure
			if err := validat.validatePayload(envelope.Type, itemMap); err != nil {
				s.Logger.Warn("Skipping malformed payload in batch", "error", err)
				continue
			}

			// Create Telemetry Model
			ts := envelope.Timestamp
			if itemTs, ok := itemMap["ts"].(float64); ok {
				ts = int64(itemTs)
			}

			t := models.Telemetry{
				DeviceID:  deviceID,
				Timestamp: ts,
				Type:      envelope.Type,
				Payload:   itemMap,
			}

			telemetryList = append(telemetryList, t)
			latestReading = t // Keep track of the last one for state update
		}

		// Save the whole tray at once
		if len(telemetryList) > 0 {
			if err := s.TelemetryStore.SaveTelemetryBatch(ctx, telemetryList); err != nil {
				s.Logger.Error("Failed to save batch telemetry", "error", err)
				return err
			}
			
			// OPTIMIZATION: Update state only once with the latest reading
			return s.StateStore.UpdateFromTelemetry(ctx, latestReading)
		}
		return nil
	}

	// --- PATH B: SINGLE PROCESSING ---
	data := models.Telemetry{
		DeviceID:  envelope.DeviceID,
		Timestamp: envelope.Timestamp,
		Type:      envelope.Type,
		Payload:   envelope.Payload,
	}

	s.Logger.Info("Saving Single Telemetry", "device_id", deviceID)

	if err := s.TelemetryStore.SaveTelemetry(ctx, data); err != nil {
		s.Logger.Error("Failed to save telemetry", "error", err)
		return err
	}

	return s.StateStore.UpdateFromTelemetry(ctx, data)
}

func (s *Service) handleAlert(ctx context.Context, deviceID string, envelope models.MQTTEnvelope) error {
	severity, _ := envelope.Payload["severity"].(string)

	alert := models.Alert{
		DeviceID:  deviceID,
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

func (s *Service) logValidationError(err error, deviceID string) {
	switch {
	case errors.Is(err, validation.ErrInvalidEvent), errors.Is(err, validation.ErrInvalidEnvelope):
		s.Logger.Warn("Invalid message envelope", "error", err)
	case errors.Is(err, validation.ErrInvalidPayload):
		s.Logger.Warn("Invalid payload", "device_id", deviceID, "error", err)
	case errors.Is(err, validation.ErrInvalidTopic):
		s.Logger.Warn("Invalid topic", "error", err)
	default:
		s.Logger.Error("Unexpected validation error", "error", err)
	}
}
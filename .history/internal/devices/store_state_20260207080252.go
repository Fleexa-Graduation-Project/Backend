package devices

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/Fleexa-Graduation-Project/Backend/models"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/db"
)

type StateStore struct {
	Client    *dynamodb.Client
	TableName string
}

func NewStateStore() (*StateStore, error) {
	tableName := os.Getenv("DYNAMODB_DEVICE_STATE_TABLE")
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_DEVICE_STATE_TABLE is not set")
	}

	if db.Client == nil {
		return nil, fmt.Errorf("dynamodb client not initialized")
	}

	return &StateStore{
		Client:    db.Client,
		TableName: tableName,
	}, nil
}

func (s *StateStore) UpdateState(ctx context.Context, state models.DeviceState) error {
	state.LastUpdated = time.Now().Unix()

	item, err := attributevalue.MarshalMap(state)
	if err != nil {
		return fmt.Errorf("failed to marshal device state: %w", err)
	}

	_, err = s.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.TableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert device state: %w", err)
	}

	return nil
}

// ExtractState derives the operational state from the device type and its payload
func ExtractState(
	deviceType string,
	payload map[string]interface{},
) string {

	switch deviceType {

	case "temp-sensor":
		if temp, ok := payload["temp"].(float64);
		 ok {
			if temp > 30 {
				return "HOT"
			}
			if temp < 18 {
				return "COLD"
			}
			return "NORMAL"
		}

	case "light-sensor":
		if level, ok := payload["light_level"].(float64); 
		ok {
			if level > 600 {
				return "BRIGHT"
			}
			if level < 200 {
				return "DIM"
			}
			return "NORMAL"
		}

	case "door-actuator":
		if state, ok := payload["lock_state"].(string); 
		ok {
			return state // LOCKED or UNLOCKED
		}
	}

	return "UNKNOWN"
}


func EvaluateHealth(
	deviceType string,
	operationalState string,
) string {

	switch deviceType {

	case "temp-sensor":
		switch operationalState {
		case "HOT":
			return "DEGRADED"
		case "COLD", "NORMAL":
			return "HEALTHY"
		}

	case "gas-sensor":
		switch operationalState {
		case "DANGER":
			return "CRITICAL"
		case "WARNING":
			return "DEGRADED"
		case "SAFE":
			return "HEALTHY"
		}

	case "door-actuator":
		return "HEALTHY"
	}

	return "DEGRADED"
}

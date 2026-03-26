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

const (
	OfflineLimit = 2 * time.Minute
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
	now := time.Now().Unix()

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]types.AttributeValue{
			"device_id": &types.AttributeValueMemberS{Value: state.DeviceID},
		},
		UpdateExpression: aws.String(`
			SET 
				#type = :type,
				#status = :status,
				operational_state = :op_state,
				health = :health,
				last_seen_at = :last_seen,
				updated_at = :updated_at
		`),
		ExpressionAttributeNames: map[string]string{
			"#type":   "type",
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":type":        &types.AttributeValueMemberS{Value: state.Type},
			":status":      &types.AttributeValueMemberS{Value: state.Status},
			":op_state":    &types.AttributeValueMemberS{Value: state.OperationalState},
			":health":      &types.AttributeValueMemberS{Value: state.Health},
			":last_seen":   &types.AttributeValueMemberN{Value: fmt.Sprint(state.LastSeenAt)},
			":updated_at":  &types.AttributeValueMemberN{Value: fmt.Sprint(now)},
		},
	} 

	_, err := s.Client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update device state: %w", err)
	}

	return nil
}


//Extracting operational state from the device type and its payload
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


func ConnectionStatus(lastSeenAt int64) string {
	if time.Since(time.Unix(lastSeenAt, 0)) > OfflineLimit {
		return "OFFLINE"
	}
	return "ONLINE"
}

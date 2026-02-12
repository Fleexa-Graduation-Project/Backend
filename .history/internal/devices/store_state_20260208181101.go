package devices

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
//saving state
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
func (s *StateStore) UpdateFromTelemetry(ctx context.Context,state models.Telemetry,) error {

	now := time.Now().Unix()
	
	opState, health := ExtractState(state.Type, state.Payload)

	status := "ONLINE"
	if ConnectionStatus(state.Timestamp) == "OFFLINE" {
		status = "OFFLINE"
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]types.AttributeValue{
			"device_id": &types.AttributeValueMemberS{
				Value: t.DeviceID,
			},
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
			":type":       &types.AttributeValueMemberS{Value: t.Type},
			":status":     &types.AttributeValueMemberS{Value: status},
			":op_state":   &types.AttributeValueMemberS{Value: opState},
			":health":     &types.AttributeValueMemberS{Value: health},
			":last_seen":  &types.AttributeValueMemberN{Value: fmt.Sprint(t.Timestamp)},
			":updated_at": &types.AttributeValueMemberN{Value: fmt.Sprint(now)},
		},
	}

	_, err := s.Client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update device state: %w", err)
	}

	return nil
}



//Extracting operational state and health from the device type and its payload
func ExtractState(deviceType string,payload map[string]interface{},) (string, string)  {

	deviceRules, ok := Rules[deviceType]
opState := "UNKNOWN"
	health := "DEGRADED"

	if ok {
		opState = deviceRules.ExtractOperational(payload)
		health = deviceRules.EvaluateHealth(opState)
	}

	return opState, health
}



func ConnectionStatus(lastSeenAt int64) string {
	if time.Since(time.Unix(lastSeenAt, 0)) > OfflineLimit {
		return "OFFLINE"
	}
	return "ONLINE"
}

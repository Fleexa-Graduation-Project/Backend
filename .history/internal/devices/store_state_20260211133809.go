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

func (s *StateStore) UpdateFromTelemetry(ctx context.Context,tel models.Telemetry,) error {

	now := time.Now().Unix()
	
	opState, health := ExtractState(tel.Type, tel.Payload)
	status := "ONLINE"

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]types.AttributeValue{
			"device_id": &types.AttributeValueMemberS{
				Value: tel.DeviceID,
			},
		},
		ConditionExpression: aws.String(
            "attribute_not_exists(last_seen_at) OR last_seen_at <= :last_seen",
        ),
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
			":type":       &types.AttributeValueMemberS{Value: tel.Type},
			":status":     &types.AttributeValueMemberS{Value: status},
			":op_state":   &types.AttributeValueMemberS{Value: opState},
			":health":     &types.AttributeValueMemberS{Value: health},
			":last_seen":  &types.AttributeValueMemberN{Value: fmt.Sprint(tel.Timestamp)},
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

func (s *StateStore) UpdateHeartbeat(ctx context.Context,deviceID string,) error {
	if
    now := time.Now().Unix()

    input := &dynamodb.UpdateItemInput{
        TableName: aws.String(s.TableName),
        Key: map[string]types.AttributeValue{
            "device_id": &types.AttributeValueMemberS{Value: deviceID},
        },
		ConditionExpression: aws.String(
            "attribute_not_exists(last_seen_at) OR last_seen_at <= :last_seen",
        ),
        UpdateExpression: aws.String(
            "SET #status = :status, last_seen_at = :last_seen",
        ),
        ExpressionAttributeNames: map[string]string{
            "#status": "status",
        },
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":status":    &types.AttributeValueMemberS{Value: "ONLINE"},
            ":last_seen": &types.AttributeValueMemberN{Value: fmt.Sprint(now)},
        },
    }

    _, err := s.Client.UpdateItem(ctx, input)
    return err
}

func ConnectionStatus(lastSeenAt int64) string {
	if time.Since(time.Unix(lastSeenAt, 0)) > OfflineLimit {
		return "OFFLINE"
	}
	return "ONLINE"
}

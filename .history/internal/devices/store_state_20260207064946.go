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

func (s *StateStore) Upsert(ctx context.Context, state models.DeviceState) error {
	state.UpdatedAt = time.Now().Unix()

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

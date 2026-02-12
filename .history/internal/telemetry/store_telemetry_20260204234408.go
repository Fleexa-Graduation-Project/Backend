package db

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

type TelemetryStore struct {
	Client    *dynamodb.Client
	TableName string
}

func NewTelemetryStore(client *DynamoDBClient) (*TelemetryStore, error) {
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE_NAME environment variable is not set")
	}

	return &TelemetryStore{
		Client:    client.Client,
		TableName: tableName,
	}, nil
}

func (store *TelemetryStore) SaveTelemetry(ctx context.Context, data models.Telemetry) error {
	if data.ExpiresAt == 0 {
		data.ExpiresAt = time.Now().Add(7 * 24 * time.Hour).Unix()
	}

	item, err := attributevalue.MarshalMap(data)

	if err != nil {
		return fmt.Errorf("failed to marshal telemetry data: %v", err)

	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(store.TableName),
		Item:      item,
	}

	_, err = store.Client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to store data into DynamoDB: %v", err)
	}

	return nil
}
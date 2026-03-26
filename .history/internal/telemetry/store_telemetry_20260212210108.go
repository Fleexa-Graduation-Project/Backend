package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Fleexa-Graduation-Project/Backend/models"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/db"
	"github.com/aws/aws-sdk-go-v2/aws"
    
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	dynamoBatchLimit = 25 // DynamoDB BatchWriteItem hard limit
	maxRetries       = 3  // Retries for unprocessed items
)

type TelemetryStore struct {
	Client    *dynamodb.Client
	TableName string
}

// NewTelemetryStore initializes the store using the shared db.Client
func NewTelemetryStore() (*TelemetryStore, error) {
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE_NAME environment variable is not set")
	}

	// We use the global 'db.Client' we created in pkg/db/client.go
	return &TelemetryStore{
		Client:    db.Client,
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

// storing multiple telemetry records in a single DynamoDB call(max 25)
func (store *TelemetryStore) SaveTelemetryBatch(ctx context.Context, dataList []models.Telemetry) error {
	if len(dataList) == 0 {
		return nil
	}

	if len(dataList) > 25 {
		return fmt.Errorf("batch size exceeds DynamoDB limit of 25 items")
	}

	var writeRequests []types.WriteRequest
	defaultExpiry := time.Now().Add(7 * 24 * time.Hour).Unix()

	for _, data := range dataList {
		// Setting default TTL if not provided
		if data.ExpiresAt == 0 {
			data.ExpiresAt = defaultExpiry
		}

		// Marshalling the struct into DynamoDB attributes
		item, err := attributevalue.MarshalMap(data)
		if err != nil {
			return fmt.Errorf("failed to marshal batch item: %v", err)
		}

		// Wrapping the item in a PutRequest
		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: item,
			},
		})
	}

	// Executing the batch write
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			store.TableName: writeRequests,
		},
	}

	_, err := store.Client.BatchWriteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute batch write: %v", err)
	}

	return nil
}

// writeBatchWithRetry writes a single chunk and retries any UnprocessedItems
func (store *TelemetryStore) writeBatchWithRetry(ctx context.Context, requests []types.WriteRequest) error {
	pending := requests

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Exponential backoff on retries (100ms, 200ms, 400ms)
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				store.TableName: pending,
			},
		}

		output, err := store.Client.BatchWriteItem(ctx, input)
		if err != nil {
			return fmt.Errorf("batch write attempt %d failed: %w", attempt+1, err)
		}

		// Check for unprocessed items (DynamoDB throttling)
		unprocessed := output.UnprocessedItems[store.TableName]
		if len(unprocessed) == 0 {
			return nil
		}

		pending = unprocessed
	}

	return fmt.Errorf("batch write: %d items still unprocessed after %d retries", len(pending), maxRetries)
}



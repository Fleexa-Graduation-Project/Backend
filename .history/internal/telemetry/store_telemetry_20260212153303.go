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

// storing multiple telemetry records in a single DynamoDB call (max 25)
func (store *TelemetryStore) SaveTelemetryBatch(ctx context.Context, dataList []models.Telemetry) error {
	if len(dataList) == 0 {
		return nil
	}

	defaultExpiry := time.Now().Add(7 * 24 * time.Hour).Unix()

	// Split into chunks of 25 (DynamoDB BatchWriteItem limit)
	const batchSize = 25
	for i := 0; i < len(dataList); i += batchSize {
		end := i + batchSize
		if end > len(dataList) {
			end = len(dataList)
		}
		chunk := dataList[i:end]

		// Prepare write requests for this chunk
		var writeRequests []types.WriteRequest
		for _, data := range chunk {
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

		// Process this chunk with retry logic
		if err := store.processBatchWithRetry(ctx, writeRequests); err != nil {
			return err
		}
	}

	return nil
}

// processBatchWithRetry handles a batch write with exponential backoff retry for unprocessed items
func (store *TelemetryStore) processBatchWithRetry(ctx context.Context, writeRequests []types.WriteRequest) error {
	const maxRetries = 5
	backoff := 100 * time.Millisecond

	requestItems := map[string][]types.WriteRequest{
		store.TableName: writeRequests,
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if len(requestItems[store.TableName]) == 0 {
			// All items processed successfully
			return nil
		}

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: requestItems,
		}

		result, err := store.Client.BatchWriteItem(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to execute batch write (attempt %d/%d): %v", attempt+1, maxRetries, err)
		}

		// Check for unprocessed items
		if len(result.UnprocessedItems) == 0 {
			// All items processed successfully
			return nil
		}

		// If there are unprocessed items and we have retries left, prepare for retry
		if attempt < maxRetries-1 {
			requestItems = result.UnprocessedItems
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		} else {
			// Last attempt and still have unprocessed items
			return fmt.Errorf("failed to process %d items after %d attempts", len(result.UnprocessedItems[store.TableName]), maxRetries)
		}
	}

	return nil
}

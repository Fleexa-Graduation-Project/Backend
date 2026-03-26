package telemetry

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "github.com/Fleexa-Graduation-Project/Backend/models"
    "github.com/Fleexa-Graduation-Project/Backend/pkg/db"
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

//storing multiple telemetry records with automatic chunking and retry logic
func (store *TelemetryStore) SaveTelemetryBatch(ctx context.Context, dataList []models.Telemetry) error {
	if len(dataList) == 0 {
		return nil
	}

	// Split data into chunks of 25 (DynamoDB's batch write limit)
	const chunkSize = 25
	for i := 0; i < len(dataList); i += chunkSize {
		end := i + chunkSize
		if end > len(dataList) {
			end = len(dataList)
		}
		chunk := dataList[i:end]

		// Process each chunk with retry logic
		if err := store.processBatchChunk(ctx, chunk); err != nil {
			return fmt.Errorf("failed to process batch chunk (items %d-%d): %v", i, end-1, err)
		}
	}

	return nil
}

// processBatchChunk handles a single chunk of up to 25 items with retry logic
func (store *TelemetryStore) processBatchChunk(ctx context.Context, dataList []models.Telemetry) error {
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

	// Retry logic with exponential backoff for unprocessed items
	const maxRetries = 5
	backoff := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		if len(writeRequests) == 0 {
			break
		}

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				store.TableName: writeRequests,
			},
		}

		output, err := store.Client.BatchWriteItem(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to execute batch write: %v", err)
		}

		// Check for unprocessed items
		unprocessedItems, exists := output.UnprocessedItems[store.TableName]
		if !exists || len(unprocessedItems) == 0 {
			// All items processed successfully
			return nil
		}

		// Prepare for retry with unprocessed items
		writeRequests = unprocessedItems

		// Wait before retrying (exponential backoff)
		if attempt < maxRetries-1 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled while retrying unprocessed items")
			case <-time.After(backoff):
				backoff *= 2 // Exponential backoff
			}
		}
	}

	// If we still have unprocessed items after max retries
	if len(writeRequests) > 0 {
		return fmt.Errorf("failed to process %d items after %d retries", len(writeRequests), maxRetries)
	}

	return nil
}
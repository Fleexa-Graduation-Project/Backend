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

const (
    dynamoBatchLimit = 25 // DynamoDB BatchWriteItem hard limit
    maxRetries       = 3  // retries for unprocessed items
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

// SaveTelemetryBatch stores telemetry records using DynamoDB BatchWriteItem.
// It automatically chunks into groups of 25 (DynamoDB limit) and retries
// any unprocessed items with exponential backoff.
func (store *TelemetryStore) SaveTelemetryBatch(ctx context.Context, dataList []models.Telemetry) error {
    if len(dataList) == 0 {
        return nil
    }

    defaultExpiry := time.Now().Add(7 * 24 * time.Hour).Unix()

    // Marshal all items upfront so we fail fast on bad data
    allRequests := make([]types.WriteRequest, 0, len(dataList))
    for i := range dataList {
        if dataList[i].ExpiresAt == 0 {
            dataList[i].ExpiresAt = defaultExpiry
        }

        item, err := attributevalue.MarshalMap(dataList[i])
        if err != nil {
            return fmt.Errorf("failed to marshal batch item %d: %v", i, err)
        }

        allRequests = append(allRequests, types.WriteRequest{
            PutRequest: &types.PutRequest{
                Item: item,
            },
        })
    }

    // Process in chunks of 25 (DynamoDB hard limit)
    for start := 0; start < len(allRequests); start += dynamoBatchLimit {
        end := start + dynamoBatchLimit
        if end > len(allRequests) {
            end = len(allRequests)
        }

        if err := store.writeBatchWithRetry(ctx, allRequests[start:end]); err != nil {
            return fmt.Errorf("batch chunk [%d:%d] failed: %w", start, end, err)
        }
    }

    return nil
}

// writeBatchWithRetry writes a single chunk (≤25 items) and retries any
// UnprocessedItems with exponential backoff.
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

        output, err := store.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
            RequestItems: map[string][]types.WriteRequest{
                store.TableName: pending,
            },
        })
        if err != nil {
            return fmt.Errorf("batch write attempt %d failed: %w", attempt+1, err)
        }

        // Check for unprocessed items (DynamoDB throttling / capacity)
        unprocessed := output.UnprocessedItems[store.TableName]
        if len(unprocessed) == 0 {
            return nil
        }

        pending = unprocessed
    }

    return fmt.Errorf("batch write: %d items still unprocessed after %d retries", len(pending), maxRetries)
}
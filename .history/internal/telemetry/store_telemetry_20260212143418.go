package telemetry

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

// SaveBatch stores up to 25 telemetry records in one go
func (s *TelemetryStore) SaveBatch(ctx context.Context, items []models.Telemetry) error {
	if len(items) == 0 {
		return nil
	}

	writeRequests := make([]types.WriteRequest, 0, len(items))

	for _, item := range items {
		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			return err
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: av,
			},
		})
	}

	// DynamoDB BatchWriteItem has a hard limit of 25 items per request
	// For your project, we'll assume the batch is <= 25. 
	// If it's more, we'd need a loop to chunk them.
	_, err := db.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			"Fleexa_Telemetry": writeRequests,
		},
	})

	return err
}
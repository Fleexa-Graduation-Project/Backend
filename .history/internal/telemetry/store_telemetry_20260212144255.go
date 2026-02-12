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

//storing multiple telemetry records in a single DynamoDB call (max 25)
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

		// Wraping the item in a PutRequest
		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: item,
			},
		})
	}

	// Execute the batch write
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
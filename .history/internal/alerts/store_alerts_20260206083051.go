
package alerts

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/Fleexa-Graduation-Project/Backend/models"
	"github.com/Fleexa-Graduation-Project/Backend/pkg/db"
)

type AlertStore struct {
	Client    *dynamodb.Client
	TableName string
}

// NewAlertStore initializes the alert store
func NewAlertStore() (*AlertStore, error) {
	tableName := os.Getenv("DYNAMODB_ALERTS_TABLE")
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_ALERTS_TABLE environment variable is not set")
	}

	if db.Client == nil {
		return nil, fmt.Errorf("dynamodb client is not initialized")
	}

	return &AlertStore{
		Client:    db.Client,
		TableName: tableName,
	}, nil
}

func (store *AlertStore) SaveAlert(ctx context.Context, alert models.Alert) error {
	if alert.ExpiresAt == 0 {
		alert.ExpiresAt = time.Now().Add(30 * 24 * time.Hour).Unix()
	}

	item, err := attributevalue.MarshalMap(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(store.TableName),
		Item:      item,
	}

	_, err = store.Client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to store alert in dynamodb: %w", err)
	}

	return nil
}


func (store *AlertStore) GetAlertsBySeverity(
	ctx context.Context,
	severity string,
) ([]models.Alert, error) {

	input := &dynamodb.QueryInput{
		TableName: aws.String(store.TableName),

		// 🔑 This is the key part
		IndexName: aws.String("SeverityIndex"),

		KeyConditionExpression: aws.String("severity = :severity"),

		ExpressionAttributeValues: map[string]types.AttributeValue{
			":severity": &types.AttributeValueMemberS{
				Value: severity,
			},
		},

		// Optional but very important
		ScanIndexForward: aws.Bool(false), // newest first

		Limit: aws.Int32(limit),
	}

	result, err := store.Client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts by severity: %w", err)
	}

	var alerts []models.Alert
	err = attributevalue.UnmarshalListOfMaps(result.Items, &alerts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal alerts: %w", err)
	}

	return alerts, nil
}

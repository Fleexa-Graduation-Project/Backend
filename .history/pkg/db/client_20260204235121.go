package db

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBClient struct {
	Client *dynamodb.Client
}

var (
	instance *DynamoDBClient
	once     sync.Once //logs to aws only once
)

// new client initializes the connection to db
func NewDynamoDBClient(ctx context.Context) (*DynamoDBClient, error) {
	var initErr error

	once.Do(func() {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			initErr = fmt.Errorf("unable to load SDK config: %v", err)
			return
		}

		//Create the Client
		instance = &DynamoDBClient{
			Client: dynamodb.NewFromConfig(cfg),
		}
		log.Println("DynamoDB Connection Established")
	})

	if initErr != nil {
		return nil, initErr
	}
	return instance, nil
}
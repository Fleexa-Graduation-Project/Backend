package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"

	"github.com/Fleexa-Graduation-Project/Backend/internal/api"
)

func main() {
	server := api.NewServer()

	lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return server.LambdaHandler().ProxyWithContext(ctx, req)
	})
}
package main

import (
	"context"
	"log/slog"

	https://github.com/Fleexa-Graduation-Project/Backend
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	log *slog.Logger
)

func init() {
	log = logger.InitLogger()
	log.Info("API Service: Cold Start Initialization")
}

// HandleRequest is the entry point for API Gateways
func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	
	// Log the incoming HTTP req path ex: "/login"
	log.Info("API Request Received", 
		"path", request.Path,
		"method", request.HTTPMethod,
	)

	// TODO: will add the "Router" here to send "/login" to the Auth function later

	// Return a simple "200 OK" response
	return events.APIGatewayProxyResponse{
		Body:       `{"message": "Hello from Fleexa API"}`,
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
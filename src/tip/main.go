package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response, _ := json.Marshal(e.PathParameters)
	return events.APIGatewayProxyResponse{
		Body:       string(response[:]),
		StatusCode: 201,
	}, nil
}

func main() {
	lambda.Start(Handler)
}

package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       "Hello from Build Breaker!",
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}

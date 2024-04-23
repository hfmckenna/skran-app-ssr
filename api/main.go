package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	query := req.QueryStringParameters["q"]
	return events.APIGatewayProxyResponse{StatusCode: 200, Headers: map[string]string{"Content-Type": "text/html"}, Body: fmt.Sprintf("<span>%s</span>", query)}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

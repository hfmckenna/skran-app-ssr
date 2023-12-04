package main

import (
	"bytes"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"skran-app-ssr/api"
)

type MyEvent struct {
	Name string `json:"name"`
}

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var buf bytes.Buffer
	api.Test(&buf)
	s := buf.String()
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: s}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

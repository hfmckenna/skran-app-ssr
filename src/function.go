package main

import (
	"bytes"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"skran-app-ssr/api"
)

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var buf bytes.Buffer
	api.Home(&buf)
	s := buf.String()
	return events.APIGatewayProxyResponse{StatusCode: 200, Headers: map[string]string{"Content-Type": "text/html"}, Body: s}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

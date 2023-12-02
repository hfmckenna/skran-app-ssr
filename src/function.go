package main

import (
	"bytes"
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"skran-app-ssr/api"
)

type MyEvent struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	var buf bytes.Buffer
	api.Test(&buf)
	s := buf.String()
	return s, nil
}

func main() {
	lambda.Start(HandleRequest)
}

package main

import (
	"bytes"
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"html/template"
	"log"
)

type MyEvent struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	templ, err := template.New("").Parse("Hi there")
	var buf bytes.Buffer
	if err != nil {
		log.Fatal(err)
	}
	err = templ.Execute(&buf, "")
	if err != nil {
		log.Fatal(err)
	}
	s := buf.String()
	return s, nil
}

func main() {
	lambda.Start(HandleRequest)
}

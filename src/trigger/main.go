package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(uow events.DynamoDBEvent) {
	records := uow.Records
	stream := sliceToStream(records)
	for v := range stream {
		fmt.Println(v)
	}
}

func main() {
	lambda.Start(HandleRequest)
}

func sliceToStream(slice []events.DynamoDBEventRecord) <-chan events.DynamoDBEventRecord {
	stream := make(chan events.DynamoDBEventRecord)
	go func() {
		for _, v := range slice {
			stream <- v
		}
		close(stream)
	}()
	return stream
}

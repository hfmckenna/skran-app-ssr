package trigger

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodbstreams/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"os"
	models "skran-app-ssr/src"
)

func HandleRequest(uow events.DynamoDBEvent) {
	region := os.Getenv("AWS_REGION")
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := dynamodb.New(sess)
	records, err := models.FromDynamoDBEvent(uow)
	if err != nil {
		return
	}
	stream := sliceToStream(records)
	for v := range stream {
		if v.EventName == "INSERT" {
			var title string
			stringErr := attributevalue.Unmarshal(v.Dynamodb.NewImage["Id"], &title)
			if stringErr != nil {
				return
			}
			svc.PutItemRequest(&dynamodb.PutItemInput{
				TableName: aws.String("SkranAppTable"),
				Item: map[string]*dynamodb.AttributeValue{
					"Primary": {
						S: aws.String(title),
					},
				},
			})
		}
	}
}

func main() {
	lambda.Start(HandleRequest)
}

func sliceToStream(slice []types.Record) <-chan types.Record {
	stream := make(chan types.Record)
	go func() {
		for _, v := range slice {
			stream <- v
		}
		close(stream)
	}()
	return stream
}

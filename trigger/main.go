package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodbstreams/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"os"
	"skran-app-ssr/models"
	"strings"
)

func HandleRequest(uow events.DynamoDBEvent) (events.DynamoDBEvent, error) {
	region := os.Getenv("AWS_REGION")
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := dynamodb.New(sess)
	records, err := FromDynamoDBEvent(uow)
	if err != nil {
		log.Fatalln("error 1:", err)
	}
	stream := sliceToStream(records)
	for v := range stream {
		if v.EventName == "REMOVE" {
			var title string
			var id string
			var components []models.Component
			err := attributevalue.Unmarshal(v.Dynamodb.OldImage["Title"], &title)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.OldImage["Id"], &id)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.OldImage["Components"], &components)
			if err != nil {
				log.Fatal(err)
			}
			writeRequest := make(map[string][]*dynamodb.WriteRequest)
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					writeRequest["SkranAppTable"] = append(writeRequest["SkranAppTable"], &dynamodb.WriteRequest{
						PutRequest: &dynamodb.PutRequest{
							Item: map[string]*dynamodb.AttributeValue{
								"Primary": {
									S: aws.String("SEARCH#" + upperSnakeCase(getFirstChar(ingredient.Title))),
								},
								"Sort": {
									S: aws.String("SEARCH#" + upperSnakeCase(ingredient.Title)),
								},
								"Title": {
									S: aws.String(ingredient.Title),
								},
								"Recipe Title": {
									S: aws.String(title),
								},
								"Recipe Id": {
									S: aws.String(id),
								},
								"Type": {
									S: aws.String("SEARCH"),
								},
								"Deleted": {
									BOOL: aws.Bool(true),
								},
							},
						},
					})
				}
			}
			req, resp := svc.BatchWriteItemRequest(&dynamodb.BatchWriteItemInput{
				RequestItems: writeRequest,
			})
			err = req.Send()
			println(resp)
			if err != nil {
				log.Fatal(err)
			}
		}
		if v.EventName == "INSERT" {
			var title string
			var id string
			var components []models.Component
			err := attributevalue.Unmarshal(v.Dynamodb.NewImage["Title"], &title)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.NewImage["Id"], &id)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.NewImage["Components"], &components)
			if err != nil {
				log.Fatal(err)
			}
			writeRequest := make(map[string][]*dynamodb.WriteRequest)
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					writeRequest["SkranAppTable"] = append(writeRequest["SkranAppTable"], &dynamodb.WriteRequest{
						PutRequest: &dynamodb.PutRequest{
							Item: map[string]*dynamodb.AttributeValue{
								"Primary": {
									S: aws.String("SEARCH#" + upperSnakeCase(getFirstChar(ingredient.Title))),
								},
								"Sort": {
									S: aws.String("SEARCH#" + upperSnakeCase(ingredient.Title)),
								},
								"Title": {
									S: aws.String(ingredient.Title),
								},
								"Recipe Title": {
									S: aws.String(title),
								},
								"Recipe Id": {
									S: aws.String(id),
								},
								"Type": {
									S: aws.String("SEARCH"),
								},
								"Deleted": {
									BOOL: aws.Bool(false),
								},
							},
						},
					})
				}
			}
			req, resp := svc.BatchWriteItemRequest(&dynamodb.BatchWriteItemInput{
				RequestItems: writeRequest,
			})
			err = req.Send()
			println(resp)
			if err != nil {
				log.Fatal(err)
			}
		}
		if v.EventName == "MODIFY" {
			var title string
			var id string
			var components []models.Component
			err := attributevalue.Unmarshal(v.Dynamodb.OldImage["Title"], &title)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.OldImage["Id"], &id)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.OldImage["Components"], &components)
			if err != nil {
				log.Fatal(err)
			}
			writeRequest := make(map[string][]*dynamodb.WriteRequest)
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					writeRequest["SkranAppTable"] = append(writeRequest["SkranAppTable"], &dynamodb.WriteRequest{
						PutRequest: &dynamodb.PutRequest{
							Item: map[string]*dynamodb.AttributeValue{
								"Primary": {
									S: aws.String("SEARCH#" + upperSnakeCase(getFirstChar(ingredient.Title))),
								},
								"Sort": {
									S: aws.String("SEARCH#" + upperSnakeCase(ingredient.Title)),
								},
								"Title": {
									S: aws.String(ingredient.Title),
								},
								"Recipe Title": {
									S: aws.String(title),
								},
								"Recipe Id": {
									S: aws.String(id),
								},
								"Type": {
									S: aws.String("SEARCH"),
								},
								"Deleted": {
									BOOL: aws.Bool(true),
								},
							},
						},
					})
				}
			}
			req, resp := svc.BatchWriteItemRequest(&dynamodb.BatchWriteItemInput{
				RequestItems: writeRequest,
			})
			err = req.Send()
			println(resp)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.NewImage["Title"], &title)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.NewImage["Id"], &id)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.NewImage["Components"], &components)
			if err != nil {
				log.Fatal(err)
			}
			writeRequest = make(map[string][]*dynamodb.WriteRequest)
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					writeRequest["SkranAppTable"] = append(writeRequest["SkranAppTable"], &dynamodb.WriteRequest{
						PutRequest: &dynamodb.PutRequest{
							Item: map[string]*dynamodb.AttributeValue{
								"Primary": {
									S: aws.String("SEARCH#" + upperSnakeCase(getFirstChar(ingredient.Title))),
								},
								"Sort": {
									S: aws.String("SEARCH#" + upperSnakeCase(ingredient.Title)),
								},
								"Title": {
									S: aws.String(ingredient.Title),
								},
								"Recipe Title": {
									S: aws.String(title),
								},
								"Recipe Id": {
									S: aws.String(id),
								},
								"Type": {
									S: aws.String("SEARCH"),
								},
								"Deleted": {
									BOOL: aws.Bool(false),
								},
							},
						},
					})
				}
			}
			req, resp = svc.BatchWriteItemRequest(&dynamodb.BatchWriteItemInput{
				RequestItems: writeRequest,
			})
			err = req.Send()
			println(resp)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return events.DynamoDBEvent{}, nil
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

func upperSnakeCase(s string) string {
	upper := strings.ToUpper(s)
	snake := strings.ReplaceAll(upper, " ", "_")
	return snake
}

func getFirstChar(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(s[0])
}

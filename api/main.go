package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/jsii-runtime-go"
	"log"
	"os"
	"skran-app-ssr/models"
	"strings"
)

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	region := os.Getenv("AWS_REGION")
	query := req.QueryStringParameters["q"]
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	ddb := dynamodb.New(sess)
	response := ""
	if len(query) > 3 {
		result, err := ddb.Query(&dynamodb.QueryInput{
			TableName:              aws.String("SkranAppTable"),
			KeyConditionExpression: jsii.String("#pk = :char and begins_with(#sk, :query)"),
			ExpressionAttributeNames: map[string]*string{
				"#pk": jsii.String("Primary"),
				"#sk": jsii.String("Sort"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":char":  {S: jsii.String(strings.ToUpper(getFirstChar(query)))},
				":query": {S: jsii.String(fmt.Sprintf("SEARCH#%s", upperSnakeCase(query)))},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		html := make([]string, len(result.Items))
		for i, item := range result.Items {
			// Use fmt.Sprintf to interpolate string
			searchItem := models.SearchItem{}
			err = dynamodbattribute.UnmarshalMap(item, &searchItem)
			if err != nil {
				panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
			}
			html[i] = fmt.Sprintf("<span>%s</span>", searchItem.Title)
		}
		response = strings.Join(html, "")
	}
	return events.APIGatewayProxyResponse{StatusCode: 200, Headers: map[string]string{"Content-Type": "text/html"}, Body: response}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

func getFirstChar(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(s[0])
}

func upperSnakeCase(s string) string {
	upper := strings.ToUpper(s)
	snake := strings.ReplaceAll(upper, " ", "_")
	return snake
}

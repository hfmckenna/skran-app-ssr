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

var sess *session.Session
var ddb *dynamodb.DynamoDB
var region string

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	query := upperSnakeCase(req.QueryStringParameters["q"])
	find := upperSnakeCase(req.QueryStringParameters["find"]) + "#"
	response := ""
	if len(query) > 2 && len(query) < 20 {
		response = queryDynamo(query)
	}
	if len(find) > 4 {
		response = queryDynamo(find)
	}
	return events.APIGatewayProxyResponse{StatusCode: 200, Headers: map[string]string{"Content-Type": "text/html"}, Body: response}, nil
}

func main() {
	region = os.Getenv("AWS_REGION")
	sess = session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	ddb = dynamodb.New(sess)
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

func queryDynamo(query string) string {
	response := ""
	result, err := ddb.Query(&dynamodb.QueryInput{
		TableName:              aws.String("SkranAppTable"),
		KeyConditionExpression: jsii.String("#pk = :char and begins_with(#sk, :query)"),
		ExpressionAttributeNames: map[string]*string{
			"#pk": jsii.String("Primary"),
			"#sk": jsii.String("Sort"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":char":  {S: jsii.String("SEARCH#" + getFirstChar(query))},
			":query": {S: jsii.String(query)},
		},
	})
	if result.Items != nil {
		html := make([]string, len(result.Items))
		for i, item := range result.Items {
			searchItem := models.SearchItem{}
			err = dynamodbattribute.UnmarshalMap(item, &searchItem)
			if err != nil {
				panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
			}
			html[i] = fmt.Sprintf("<button hx-get=\"/v1/search\" name=\"find\" hx-target=\"#search-results\" value=\"%s\">%s</button>", searchItem.Title, searchItem.Title)
		}
		response = strings.Join(dedupeStrings(html), "")
	} else if err != nil {
		log.Fatal(err)
	}
	return response
}

func dedupeStrings(input []string) []string {
	// Create a map to track unique strings.
	seen := make(map[string]struct{})
	// Create a slice to store the result.
	var result []string
	// Loop over each string in the input slice.
	for _, str := range input {
		// Check if the string is already in the map.
		if _, exists := seen[str]; !exists {
			// If the string is not in the map, add it to the result slice and mark it as seen.
			result = append(result, str)
			seen[str] = struct{}{}
		}
	}
	return result
}

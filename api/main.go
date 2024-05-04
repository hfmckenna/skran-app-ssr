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
		result, err := queryDynamo(query)
		if err != nil {
			log.Fatal(err)
		}
		uniqueItems := dedupeSearch(result)
		html := make([]string, len(uniqueItems))
		for i, item := range uniqueItems {
			html[i] = fmt.Sprintf("<button hx-get=\"/v1/search\" name=\"find\" hx-target=\"#search-results\" value=\"%s\">%s</button>", item.Title, item.Title)
		}
		response = strings.Join(html, "\n")
	}
	if len(find) > 4 {
		result, err := queryDynamo(find)
		if err != nil {
			log.Fatal(err)
		}
		uniqueItems := dedupeSearch(result)
		html := make([]string, len(uniqueItems))
		for i, item := range uniqueItems {
			html[i] = fmt.Sprintf("<div><a href=\"/recipes/%s\">%s</a></div>", item.RecipeTitle, item.RecipeId)
		}
		response = strings.Join(html, "\n")
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

func queryDynamo(query string) ([]models.SearchItem, error) {
	var items []models.SearchItem
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
		for _, item := range result.Items {
			searchItem := models.SearchItem{}
			err = dynamodbattribute.UnmarshalMap(item, &searchItem)
		}
	}
	return items, err
}

func dedupeRecipes(items []models.SearchItem) []models.SearchItem {
	var deduped []models.SearchItem
	for _, item := range items {
		seen := make(map[string]string)
		if _, exists := seen[item.RecipeId]; !exists {
			deduped = append(deduped, item)
			seen[item.RecipeId] = item.RecipeId
		}
	}
	return deduped
}

func dedupeSearch(items []models.SearchItem) []models.SearchItem {
	var deduped []models.SearchItem
	for _, item := range items {
		seen := make(map[string]string)
		if _, exists := seen[item.Title]; !exists {
			deduped = append(deduped, item)
			seen[item.Title] = item.Title
		}
	}
	return deduped
}

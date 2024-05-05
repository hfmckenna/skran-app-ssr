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
	ingredient := req.QueryStringParameters["ingredient"]
	find := upperSnakeCase(req.QueryStringParameters["find"]) + "#"
	headers := map[string]string{"Content-Type": "text/html"}
	response := ""
	if len(ingredient) > 0 {
		response = fmt.Sprintf("<input type=\"text\" name=\"find\" hx-trigger=\"load\" hx-get=\"/v1/search\" hx-target=\"#active-search\" value=\"%s\" readonly />", ingredient)
	}
	if len(query) > 2 && len(query) < 20 {
		result, err := queryDynamo(query)
		if err != nil {
			log.Fatal(err)
		}
		uniqueItems := dedupeSearch(result)
		html := make([]string, len(uniqueItems))
		for i, item := range uniqueItems {
			html[i] = fmt.Sprintf("<button hx-get=\"/v1/search\" name=\"ingredient\" hx-target=\"#existing-searches\" value=\"%s\" hx-on:click=\"const search = document.getElementById('ingredient');search.value = '';search.dispatchEvent(new Event('keyup'));\">%s</button>", item.Title, item.Title)
		}
		response = strings.Join(html, "\n")
	}
	if len(find) > 4 {
		result, err := queryDynamo(find)
		if err != nil {
			log.Fatal(err)
		}
		uniqueItems := dedupeRecipes(result)
		html := make([]string, len(uniqueItems))
		for i, item := range uniqueItems {
			html[i] = fmt.Sprintf("<div><a href=\"/recipes/%s\">%s</a></div>", item.RecipeId, item.RecipeTitle)
		}
		response = strings.Join(html, "\n")
		response = "<div>" + response + "</div>"
	}
	return events.APIGatewayProxyResponse{StatusCode: 200, Headers: headers, Body: response}, nil
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
	if result != nil && result.Items != nil {
		for _, item := range result.Items {
			searchItem := models.SearchItem{}
			err = dynamodbattribute.UnmarshalMap(item, &searchItem)
			items = append(items, searchItem)
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
		seen := make(map[string]struct{})
		if _, exists := seen[item.Title]; !exists {
			deduped = append(deduped, item)
			seen[item.Title] = struct{}{}
		}
	}
	return deduped
}

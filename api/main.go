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
	searches := req.MultiValueQueryStringParameters["search"]
	searches = make([]string, len(searches))
	for i, search := range searches {
		searches[i] = upperSnakeCase(search)
	}
	response := ""
	if len(query) > 3 && len(query) < 30 && len(searches) == 0 {
		response = queryDynamo(query)
	}
	if len(searches) > 0 && len(searches) < 6 && len(query) == 0 {
		response = batchGetDynamo(searches)
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
		IndexName:              aws.String("Secondary"),
		KeyConditionExpression: jsii.String("#pk = :char and begins_with(#sk, :query)"),
		ExpressionAttributeNames: map[string]*string{
			"#pk": jsii.String("Secondary"),
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
			html[i] = fmt.Sprintf("<button hx-get=\"/v1/search\" name=\"search\" hx-target=\"#search-results\" value=\"%s\">%s</button>", searchItem.Title, searchItem.Title)
		}
		response = strings.Join(dedupeStrings(html), "")
	} else if err != nil {
		log.Fatal(err)
	}
	return response
}

func batchGetDynamo(searches []string) string {
	keys := make([]map[string]*dynamodb.AttributeValue, len(searches))
	for i := range keys {
		keys[i] = map[string]*dynamodb.AttributeValue{
			"Primary": {S: jsii.String(searches[i])},
		}
	}

	result, err := ddb.BatchGetItem(&dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{"SkranAppTable": {
			Keys: keys,
		}},
	})
	if err != nil {
		log.Fatal(err)
	}
	tables := result.Responses
	html := make([]string, len(tables))
	for _, table := range tables {
		for i, item := range table {
			searchItem := models.SearchItem{}
			err = dynamodbattribute.UnmarshalMap(item, &searchItem)
			if err != nil {
				log.Fatal(err)
			}
			html[i] = fmt.Sprintf("<a href=\"/recipe/%s\">%s</a>", searchItem.RecipeId, searchItem.RecipeTitle)
		}
	}
	response := strings.Join(dedupeStrings(html), "")
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

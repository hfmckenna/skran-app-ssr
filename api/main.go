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
	response := ""
	if len(query) > 3 && len(query) < 30 {
		result, err := queryDynamo(query)
		if err != nil {
			log.Fatal(err)
		}
		html := make([]string, len(result.Items))
		for i, item := range result.Items {
			searchItem := models.SearchItem{}
			err = dynamodbattribute.UnmarshalMap(item, &searchItem)
			if err != nil {
				panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
			}
			html[i] = fmt.Sprintf("<button hx-get=\"/v1/search\" name=\"search\" hx-target=\"#search-results\" value=\"%s\">%s</button>", searchItem.Title, searchItem.Title)
		}
		response = strings.Join(html, "")
	}
	if len(searches) > 0 && len(searches) < 6 {
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

func queryDynamo(query string) (*dynamodb.QueryOutput, error) {
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
	return result, err
}

func batchGetDynamo(searches []string) (*dynamodb.BatchGetItemOutput, error) {
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
	return result, err
}

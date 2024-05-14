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
	find := req.MultiValueQueryStringParameters["find"]
	var upperFind []string
	for _, findQuery := range find {
		upperFind = append(upperFind, upperSnakeCase(findQuery))
	}
	headers := map[string]string{"Content-Type": "text/html", "Access-Control-Allow-Origin": "https://recipes.skran.app"}
	response := ""
	if len(ingredient) > 0 {
		response = fmt.Sprintf("<input type=\"text\" name=\"find\" hx-trigger=\"load\" hx-include=\"[name='find']\" hx-get=\"/v1/search\" hx-target=\"#active-search\" value=\"%s\" readonly />", ingredient)
	}
	if len(query) > 2 && len(query) < 20 {
		dynamoValue := Queries{value: query}
		result, err := queryDynamo(dynamoValue)
		if err != nil {
			log.Fatal(err)
		}
		uniqueItems := dedupeSearch(result)
		html := make([]string, len(uniqueItems))
		for i, item := range uniqueItems {
			html[i] = fmt.Sprintf("<button hx-get=\"/v1/search\" name=\"ingredient\" hx-target=\"#existing-searches\" hx-swap=\"beforeend\" value=\"%s\" hx-on:click=\"const search = document.getElementById('ingredient');search.value = '';search.dispatchEvent(new Event('keyup'));\">%s</button>", item.Title, item.Title)
		}
		response = strings.Join(html, "\n")
	}
	if len(upperFind) > 0 {
		dynamoValue := Queries{value: upperFind}
		result, err := queryDynamo(dynamoValue)
		if err != nil {
			log.Fatal(err)
		}
		uniqueItems := dedupeRecipes(result)
		html := make([]string, len(uniqueItems))
		for i, item := range uniqueItems {
			html[i] = fmt.Sprintf("<div><a href=\"https://recipes.skran.app/%s.html\">%s</a></div>", item.RecipeId, item.RecipeTitle)
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

type Queries struct {
	value interface{}
}

func (q *Queries) IsString() bool {
	_, ok := q.value.(string)
	return ok
}

func (q *Queries) IsSlice() bool {
	_, ok := q.value.([]string)
	return ok
}

func (q *Queries) GetString() string {
	v, _ := q.value.(string)
	return v
}

func (q *Queries) GetSlice() []string {
	v, _ := q.value.([]string)
	return v
}

func queryDynamo(query Queries) ([]models.SearchItem, error) {
	var items []models.SearchItem
	var search string
	expression := map[string]*dynamodb.AttributeValue{
		":char":        {S: jsii.String("SEARCH#" + getFirstChar(search))},
		":query":       {S: jsii.String(search)},
		":ingredients": {S: jsii.String("Recipe Ingredients")},
	}
	if query.IsString() {
		search = query.GetString() + "#"
	}
	filter := ""
	if query.IsSlice() {
		q := query.GetSlice()
		search = q[0] + "#"
		for i, item := range q[1:] {
			partialFilter := fmt.Sprintf("contains(:ingredients, :val%i)", i)
			if i > 0 {
				partialFilter = " AND " + partialFilter
			}
			key := fmt.Sprintf(":val%i", i)
			expression[key] = &dynamodb.AttributeValue{S: jsii.String(item)}
			filter = filter + partialFilter
		}
	}
	result, err := ddb.Query(&dynamodb.QueryInput{
		TableName:              aws.String("SkranAppTable"),
		KeyConditionExpression: jsii.String("#pk = :char and begins_with(#sk, :query)"),
		ExpressionAttributeNames: map[string]*string{
			"#pk": jsii.String("Primary"),
			"#sk": jsii.String("Sort"),
		},
		ExpressionAttributeValues: expression,
		FilterExpression:          jsii.String(filter),
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
	seen := make(map[string]int)
	for _, item := range items {
		if _, exists := seen[item.RecipeId]; !exists {
			deduped = append(deduped, item)
			seen[item.RecipeId] = 1
		}
	}
	return deduped
}

func dedupeSearch(items []models.SearchItem) []models.SearchItem {
	var deduped []models.SearchItem
	seen := make(map[string]int)
	for _, item := range items {
		if _, exists := seen[item.Title]; !exists {
			deduped = append(deduped, item)
			seen[item.Title] = 1
		}
	}
	return deduped
}

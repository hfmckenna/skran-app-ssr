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
	remove := upperSnakeCase(req.QueryStringParameters["remove"])
	ingredient := req.QueryStringParameters["ingredient"]
	find := req.MultiValueQueryStringParameters["find"]
	var upperFind []string
	for _, findQuery := range find {
		upperFind = append(upperFind, upperSnakeCase(findQuery))
		if remove != "" {
			upperFind = append(upperFind, remove)
		}
		upperFind = removeDuplicates(upperFind)
	}
	headers := map[string]string{"Content-Type": "text/html", "Access-Control-Allow-Origin": "https://recipes.skran.app"}
	response := "<div></div>"
	if len(ingredient) > 0 {
		response = fmt.Sprintf("<div><input type=\"text\" name=\"find\" hx-trigger=\"load\" hx-include=\"[name='find']\" hx-get=\"/v1/search\" hx-target=\"#active-search\" value=\"%s\" class=\"bg-transparent text-base lg:text-sm\" readonly /><button class=\"ml-2 text-gray-500 hover:text-gray-700\" name=\"remove\" hx-include=\"[name='find']\" hx-get=\"/v1/search\" value=\"%s\" hx-target=\"#active-search\" hx-on:click=\"this.parentNode.remove()\">X</button></div>", ingredient, ingredient)
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
			html[i] = fmt.Sprintf("<div><button class=\"text-xl lg:text-base px-4 py-2 hover:bg-gray-100\" hx-get=\"/v1/search\" name=\"ingredient\" hx-target=\"#existing-searches\" hx-swap=\"beforeend\" value=\"%s\" hx-on:click=\"const search = document.getElementById('ingredient');search.value = '';search.dispatchEvent(new Event('keyup'));\">%s</button></div>", item.Title, item.Title)
		}
		response = strings.Join(html, "\n")
	}
	if len(upperFind) > 0 && len(upperFind) < 4 {
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
	filter := ""
	expression := map[string]*dynamodb.AttributeValue{}
	if query.IsString() {
		search = query.GetString()
		expression[":char"] = &dynamodb.AttributeValue{S: jsii.String("SEARCH#" + getFirstChar(search))}
		expression[":query"] = &dynamodb.AttributeValue{S: jsii.String(search)}
	}
	if query.IsSlice() {
		q := query.GetSlice()
		search = q[0]
		expression[":char"] = &dynamodb.AttributeValue{S: jsii.String("SEARCH#" + getFirstChar(search))}
		expression[":query"] = &dynamodb.AttributeValue{S: jsii.String(search + "#")}
		for i, item := range q[1:] {
			partialFilter := fmt.Sprintf("contains(#ingredients, :val%d)", i)
			if i > 0 {
				partialFilter = " AND " + partialFilter
			}
			key := fmt.Sprintf(":val%d", i)
			expression[key] = &dynamodb.AttributeValue{S: jsii.String(item)}
			filter = filter + partialFilter
		}
	}
	input := &dynamodb.QueryInput{
		TableName:              aws.String("SkranAppTable"),
		KeyConditionExpression: jsii.String("#pk = :char and begins_with(#sk, :query)"),
		ExpressionAttributeNames: map[string]*string{
			"#pk": jsii.String("Primary"),
			"#sk": jsii.String("Sort"),
		},
		ExpressionAttributeValues: expression,
	}
	if filter != "" {
		input.ExpressionAttributeNames["#ingredients"] = jsii.String("Recipe Ingredients")
		input.FilterExpression = aws.String(filter)
	}
	result, err := ddb.Query(input)
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

func removeDuplicates(slice []string) []string {
	// Create a map to count occurrences of each string
	countMap := make(map[string]int)
	for _, str := range slice {
		countMap[str]++
	}
	// Create a new slice to hold the result
	var result []string
	for _, str := range slice {
		if countMap[str] == 1 {
			result = append(result, str)
		}
	}
	return result
}

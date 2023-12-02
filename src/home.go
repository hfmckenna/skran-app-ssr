package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"html/template"
	"log"
	"net/http"
	"os"
)

const indexPage = "templates/index.html"
const headPartial = "templates/head.html"

func home(w http.ResponseWriter, r *http.Request) {
	endpoint := os.Getenv("DYNAMO_ENDPOINT")
	assets := os.Getenv("ASSETS_DOMAIN")

	sess := session.Must(session.NewSession(&aws.Config{Endpoint: aws.String(endpoint), Region: aws.String("eu-west-1"), CredentialsChainVerboseErrors: aws.Bool(true)}))

	svc := dynamodb.New(sess)
	tmpl, _ := template.New("").ParseFiles([]string{indexPage, headPartial}...)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("SkranAppTable"),
		Key: map[string]*dynamodb.AttributeValue{
			"Primary": {
				S: aws.String("RECIPE#5678"),
			},
			"Sort": {
				S: aws.String("TITLE#CHICKEN_CURRY"),
			},
		},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	if result.Item == nil {
		log.Fatalf("No item")
		return
	}

	type Data struct {
		Item      RecipeItem
		Assets    string
		PageTitle string
	}

	data := Data{
		Assets:    assets,
		PageTitle: "Skran App",
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &data.Item)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	err = tmpl.ExecuteTemplate(w, "home", &data)
	if err != nil {
		log.Fatal(err)
	}
}

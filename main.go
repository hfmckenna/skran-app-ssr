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

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	endpoint := os.Getenv("DYNAMO_ENDPOINT")

	sess := session.Must(session.NewSession(&aws.Config{Endpoint: aws.String(endpoint), Region: aws.String("eu-west-1")}))

	svc := dynamodb.New(sess)

	const indexPage = "templates/index.html"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles(indexPage))

		result, err := svc.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String("SkranAppTable"),
			Key: map[string]*dynamodb.AttributeValue{
				"Primary": {
					S: aws.String("RECIPE#1234"),
				},
				"Sort": {
					S: aws.String("TITLE#SPAGHETTI_BOLOGNESE"),
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

		item := RecipeItem{}

		err = dynamodbattribute.UnmarshalMap(result.Item, &item)
		if err != nil {
			panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}
		err = tmpl.Execute(w, item)
		if err != nil {
			log.Fatal(err)
		}
	})

	log.Printf("Listening on port %s\n\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

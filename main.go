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

type Item struct {
	Pk   string `json:"pk"`
	Sk   string `json:"sk"`
	Test string `json:"test"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("eu-west-1")}))

	svc := dynamodb.New(sess)

	const indexPage = "templates/index.html"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles(indexPage))

		result, err := svc.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String("skran-app"),
			Key: map[string]*dynamodb.AttributeValue{
				"pk": {
					S: aws.String("test"),
				},
				"sk": {
					S: aws.String("test"),
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

		item := Item{}

		err = dynamodbattribute.UnmarshalMap(result.Item, &item)
		if err != nil {
			panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}
		tmpl.Execute(w, item)
	})

	log.Printf("Listening on port %s\n\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

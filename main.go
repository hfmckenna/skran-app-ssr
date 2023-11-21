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
	assets := os.Getenv("ASSETS_DOMAIN")
	if assets == "" {
		http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	}

	sess := session.Must(session.NewSession(&aws.Config{Endpoint: aws.String(endpoint), Region: aws.String("eu-west-1")}))

	svc := dynamodb.New(sess)

	const indexPage = "templates/index.html"
	const headPartial = "templates/head.html"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.New("").ParseFiles([]string{indexPage, headPartial}...)

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
	})

	log.Printf("Listening on port %s\n\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

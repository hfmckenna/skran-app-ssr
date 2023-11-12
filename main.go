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

// Create struct to hold info about new item
type Item struct {
	pk   string
	sk   string
	test string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	const indexPage = "templates/index.html"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles(indexPage))

		result, err := svc.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String("skran-app"),
			Key: map[string]*dynamodb.AttributeValue{
				"pk": {
					N: aws.String("test"),
				},
				"sk": {
					S: aws.String("test"),
				},
			},
		})
		if err != nil {
			log.Fatal(err)
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

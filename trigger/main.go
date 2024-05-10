package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodbstreams/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"skran-app-ssr/models"
	"strings"
)

func HandleRequest(uow events.DynamoDBEvent) (events.DynamoDBEvent, error) {
	assets := os.Getenv("ASSETS_DOMAIN")
	templates := os.Getenv("TEMPLATES")
	indexPage := "/tmp/recipe.html"
	headPartial := "/tmp/head.html"
	region := os.Getenv("AWS_REGION")
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := dynamodb.New(sess)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	client := s3.NewFromConfig(cfg)
	downloader := manager.NewDownloader(client)
	if !(fileExists(indexPage) && fileExists(headPartial)) {
		err = downloadToFile(downloader, "/tmp", templates, "recipe.html")
		err = downloadToFile(downloader, "/tmp", templates, "head.html")
	}
	if err != nil {
		log.Fatalln("error:", err)
	}
	println("yep")
	tmpl, _ := template.New("").ParseFiles([]string{indexPage, headPartial}...)
	records, err := FromDynamoDBEvent(uow)
	if err != nil {
		log.Fatalln("error 1:", err)
	}
	println("hey")
	for _, v := range records {
		if v.EventName == "REMOVE" {
			var title string
			var id string
			var components []models.Component
			err := attributevalue.Unmarshal(v.Dynamodb.OldImage["Title"], &title)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.OldImage["Id"], &id)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.OldImage["Components"], &components)
			if err != nil {
				log.Fatal(err)
			}
			var ingredients []*string
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					ingredientTitle := upperSnakeCase(ingredient.Title)
					ingredients = append(ingredients, &ingredientTitle)
				}
			}
			writeRequest := make(map[string][]*dynamodb.WriteRequest)
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					writeRequest["SkranAppTable"] = append(writeRequest["SkranAppTable"], &dynamodb.WriteRequest{
						PutRequest: &dynamodb.PutRequest{
							Item: map[string]*dynamodb.AttributeValue{
								"Primary": {
									S: aws.String("SEARCH#" + upperSnakeCase(getFirstChar(ingredient.Title))),
								},
								"Sort": {
									S: aws.String(upperSnakeCase(ingredient.Title) + "#" + upperSnakeCase(title)),
								},
								"Title": {
									S: aws.String(ingredient.Title),
								},
								"Recipe Title": {
									S: aws.String(title),
								},
								"Recipe Id": {
									S: aws.String(id),
								},
								"Type": {
									S: aws.String("SEARCH"),
								},
								"Recipe Ingredients": {
									SS: ingredients,
								},
								"Deleted": {
									BOOL: aws.Bool(true),
								},
							},
						},
					})
				}
			}
			req, resp := svc.BatchWriteItemRequest(&dynamodb.BatchWriteItemInput{
				RequestItems: writeRequest,
			})
			err = req.Send()
			println(resp)
			if err != nil {
				log.Fatal(err)
			}
		}
		if v.EventName == "INSERT" {
			var title string
			var id string
			var components []models.Component
			err := attributevalue.Unmarshal(v.Dynamodb.NewImage["Title"], &title)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.NewImage["Id"], &id)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.NewImage["Components"], &components)
			if err != nil {
				log.Fatal(err)
			}
			println("woop")
			var ingredients []*string
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					ingredientTitle := upperSnakeCase(ingredient.Title)
					ingredients = append(ingredients, &ingredientTitle)
				}
			}
			println("glab")
			writeRequest := make(map[string][]*dynamodb.WriteRequest)
			data := Data{
				Assets:    assets,
				PageTitle: title,
			}
			var buffer bytes.Buffer
			err = tmpl.Execute(&buffer, data)
			println("moov")
			if err != nil {
				log.Fatal(err)
			}
			println("PUT")
			_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket: aws.String(os.Getenv("RECIPES")),
				Key:    aws.String(id + ".html"),
				Body:   bytes.NewReader([]byte(buffer.String())),
			})
			if err != nil {
				log.Fatal(err)
			}
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					writeRequest["SkranAppTable"] = append(writeRequest["SkranAppTable"], &dynamodb.WriteRequest{
						PutRequest: &dynamodb.PutRequest{
							Item: map[string]*dynamodb.AttributeValue{
								"Primary": {
									S: aws.String("SEARCH#" + upperSnakeCase(getFirstChar(ingredient.Title))),
								},
								"Sort": {
									S: aws.String(upperSnakeCase(ingredient.Title) + "#" + upperSnakeCase(title)),
								},
								"Title": {
									S: aws.String(ingredient.Title),
								},
								"Recipe Title": {
									S: aws.String(title),
								},
								"Recipe Id": {
									S: aws.String(id),
								},
								"Type": {
									S: aws.String("SEARCH"),
								},
								"Recipe Ingredients": {
									SS: ingredients,
								},
								"Deleted": {
									BOOL: aws.Bool(false),
								},
							},
						},
					})
				}
			}
			req, resp := svc.BatchWriteItemRequest(&dynamodb.BatchWriteItemInput{
				RequestItems: writeRequest,
			})
			err = req.Send()
			println(resp)
			if err != nil {
				log.Fatal(err)
			}
		}
		if v.EventName == "MODIFY" {
			var title string
			var id string
			var components []models.Component
			err := attributevalue.Unmarshal(v.Dynamodb.OldImage["Title"], &title)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.OldImage["Id"], &id)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.OldImage["Components"], &components)
			if err != nil {
				log.Fatal(err)
			}
			var ingredients []*string
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					ingredientTitle := upperSnakeCase(ingredient.Title)
					ingredients = append(ingredients, &ingredientTitle)
				}
			}
			writeRequest := make(map[string][]*dynamodb.WriteRequest)
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					writeRequest["SkranAppTable"] = append(writeRequest["SkranAppTable"], &dynamodb.WriteRequest{
						PutRequest: &dynamodb.PutRequest{
							Item: map[string]*dynamodb.AttributeValue{
								"Primary": {
									S: aws.String("SEARCH#" + upperSnakeCase(getFirstChar(ingredient.Title))),
								},
								"Sort": {
									S: aws.String(upperSnakeCase(ingredient.Title) + "#" + upperSnakeCase(title)),
								},
								"Title": {
									S: aws.String(ingredient.Title),
								},
								"Recipe Title": {
									S: aws.String(title),
								},
								"Recipe Id": {
									S: aws.String(id),
								},
								"Type": {
									S: aws.String("SEARCH"),
								},
								"Recipe Ingredients": {
									SS: ingredients,
								},
								"Deleted": {
									BOOL: aws.Bool(true),
								},
							},
						},
					})
				}
			}
			req, resp := svc.BatchWriteItemRequest(&dynamodb.BatchWriteItemInput{
				RequestItems: writeRequest,
			})
			err = req.Send()
			println(resp)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.NewImage["Title"], &title)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.NewImage["Id"], &id)
			if err != nil {
				log.Fatal(err)
			}
			err = attributevalue.Unmarshal(v.Dynamodb.NewImage["Components"], &components)
			if err != nil {
				log.Fatal(err)
			}
			for _, component := range components {
				for _, ingredient := range component.Ingredients {
					writeRequest["SkranAppTable"] = append(writeRequest["SkranAppTable"], &dynamodb.WriteRequest{
						PutRequest: &dynamodb.PutRequest{
							Item: map[string]*dynamodb.AttributeValue{
								"Primary": {
									S: aws.String("SEARCH#" + upperSnakeCase(getFirstChar(ingredient.Title))),
								},
								"Sort": {
									S: aws.String(upperSnakeCase(ingredient.Title) + "#" + upperSnakeCase(title)),
								},
								"Title": {
									S: aws.String(ingredient.Title),
								},
								"Recipe Title": {
									S: aws.String(title),
								},
								"Recipe Id": {
									S: aws.String(id),
								},
								"Type": {
									S: aws.String("SEARCH"),
								},
								"Recipe Ingredients": {
									SS: ingredients,
								},
								"Deleted": {
									BOOL: aws.Bool(false),
								},
							},
						},
					})
				}
			}
			req, resp = svc.BatchWriteItemRequest(&dynamodb.BatchWriteItemInput{
				RequestItems: writeRequest,
			})
			err = req.Send()
			println(resp)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return events.DynamoDBEvent{}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

func upperSnakeCase(s string) string {
	upper := strings.ToUpper(s)
	snake := strings.ReplaceAll(upper, " ", "_")
	return snake
}

func getFirstChar(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(s[0])
}

func downloadToFile(downloader *manager.Downloader, targetDirectory, bucket, key string) error {
	// Create the directories in the path
	file := filepath.Join(targetDirectory, key)
	// Set up the local file
	fd, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fd.Close()
	// Download the file using the AWS SDK for Go
	fmt.Printf("Downloading s3://%s/%s to %s...\n", bucket, key, file)
	_, err = downloader.Download(context.TODO(), fd, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
	return err
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	// return false if the 'file' is a directory.
	return !info.IsDir()
}

type Data struct {
	Assets    string
	PageTitle string
}

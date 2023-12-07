package api

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
)

const indexPage = "/tmp/index.html"
const headPartial = "/tmp/head.html"

func Home(w io.Writer) {
	endpoint := os.Getenv("DYNAMO_ENDPOINT")
	assets := os.Getenv("ASSETS_DOMAIN")
	region := os.Getenv("AWS_REGION")
	templates := os.Getenv("TEMPLATES")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	sess := session.Must(session.NewSession(&aws.Config{Endpoint: aws.String(endpoint), Region: aws.String(region), CredentialsChainVerboseErrors: aws.Bool(true)}))
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	client := s3.NewFromConfig(cfg)
	downloader := manager.NewDownloader(client)
	if !(fileExists(indexPage) && fileExists(headPartial)) {
		err = downloadToFile(downloader, "/tmp", templates, "index.html")
		err = downloadToFile(downloader, "/tmp", templates, "head.html")
	}
	if err != nil {
		log.Fatalln("error:", err)
	}
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

package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/corona10/goimagehash"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

func handler(ctx context.Context, s3Event events.S3Event) {
	for _, record := range s3Event.Records {
		s3e := record.S3
		fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3e.Bucket.Name, s3e.Object.Key)

		sess := session.Must(session.NewSession())

		// download file from s3
		downloader := s3manager.NewDownloader(sess)

		dir, err := ioutil.TempDir("", s3e.Bucket.Name)
		if err != nil {
			log.Fatal(err)
		}

		filename := path.Join(dir, s3e.Object.Key)

		err = os.MkdirAll(path.Dir(filename), os.ModePerm)
		if err != nil {
			fmt.Printf("Cannot make directory path %s\n", dir)
		}

		f, err := os.Create(filename)

		if err != nil {
			fmt.Printf("failed to create file %v, %v", f.Name(), err)
			return
		}

		n, err := downloader.Download(f,
			&s3.GetObjectInput{
				Bucket: aws.String(s3e.Bucket.Name),
				Key:    aws.String(s3e.Object.Key),
			})

		if err != nil || n == 0 {
			fmt.Printf("Error %s, byte length: %v\n", err, n)
			return
		}


		pmd := populatePMD(f.Name())

		// Create ImageHash
		downloadedImage, _ := os.Open(filename)
		img1, _ := jpeg.Decode(downloadedImage)
		phash, _ := goimagehash.PerceptionHash(img1)
		pmd.PerceptualHash = phash.GetHash()

		pmd.ID = strconv.FormatUint(phash.GetHash(), 10)
		pmd.Key = strconv.Quote(s3e.Object.Key)

		// Remove the underscores from -- sometimes Mom wrote what the scene was
		pmd.ParsedName = strings.ReplaceAll(s3e.Object.Key,"_", " ")

		downloadedImage.Close()

		// rekognition bits

		rkgSvc := rekognition.New(sess)

		inputRkg := &rekognition.DetectLabelsInput{
			Image: &rekognition.Image{
				S3Object: &rekognition.S3Object{
					Bucket: aws.String(s3e.Bucket.Name),
					Name: aws.String(s3e.Object.Key),
				},
			},
		}

		result, err := rkgSvc.DetectLabels(inputRkg)

		if err != nil {
			fmt.Printf("error with DetectLabels %v\n", err)
		}

		for _, lab := range result.Labels {
			l := Labels{*lab.Name, *lab.Confidence}
			pmd.Classification.Labels = append(pmd.Classification.Labels, l)
		}

		// Save to DDB
		svc := dynamodb.New(sess)
		av, err := dynamodbattribute.MarshalMap(pmd)
		if err != nil {
			fmt.Println("Got error marshalling new pmd:")
			fmt.Println(err.Error())
			return
		}

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("photograph"),
		}

		_, err = svc.PutItem(input)
		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		f.Close()

		os.Remove(f.Name())
	}
}

func main() {
	lambda.Start(handler)
}

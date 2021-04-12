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
		fmt.Printf("filename: %s\n", filename)

		fmt.Printf("temporary directory %s\n", path.Dir(filename))
		err = os.MkdirAll(path.Dir(filename), os.ModePerm)
		if err != nil {
			fmt.Printf("Cannot make directory path %s\n", dir)
		}

		f, err := os.Create(filename)

		fmt.Printf("After create, err is %v\n", err)

		if err != nil {
			fmt.Printf("failed to create file %v, %v", f.Name(), err)
			return
		}

		fmt.Printf("About to download bucket %v, object %v to filename %v\n", s3e.Bucket.Name, s3e.Object.Key, filename)

		n, err := downloader.Download(f,
			&s3.GetObjectInput{
				Bucket: aws.String(s3e.Bucket.Name),
				Key:    aws.String(s3e.Object.Key),
			})

		if err != nil || n == 0 {
			fmt.Printf("Error %s, byte length: %v\n", err, n)
			return
		}

		fmt.Printf("downloaded %v bytes to %v\n", n, f.Name())

		pmd := populatePMD(f.Name())

		// Create ImageHash
		downloadedImage, _ := os.Open(filename)
		img1, _ := jpeg.Decode(downloadedImage)
		phash, _ := goimagehash.PerceptionHash(img1)
		pmd.PerceptualHash = phash.GetHash()

		pmd.ID = strconv.FormatUint(phash.GetHash(), 10)
		pmd.Key = s3e.Object.Key

		downloadedImage.Close()

		// rekognition bits

		rkgSvc := rekognition.New(sess)

		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Print(err)
		}

		inputRkg := &rekognition.DetectLabelsInput{
			Image: &rekognition.Image{
				Bytes: buf,
			},
		}

		result, err := rkgSvc.DetectLabels(inputRkg)

		if err != nil {
			fmt.Printf("error with DetectLabels %v\n", err)
		}

		fmt.Printf("result: %v\n", result)

		for _, lab := range result.Labels {
			fmt.Printf("label result %v\n", lab.Name)
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

		fmt.Println("Before defer close")
		f.Close()

		os.Remove(f.Name())
	}
}

func main() {
	lambda.Start(handler)
}

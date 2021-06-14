package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("no IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("non 200 Response found")
)

type Labels struct {
	Name       string
	Confidence float64
}

type PhotoMetaData struct {
	Name                string
	ParsedName          string
	Artist              string
	CaptureTime         time.Time
	CaptureYear         string
	CaptureYearMonth    string
	CaptureYearMonthDay string
	Description         string
	Caption             string
	ID                  uint64
	Height              int
	Width               int
	Classification      struct {
		Labels []Labels
	}
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession())

	_, found := request.QueryStringParameters["hello"]

	if found {
		resp, err := http.Get(DefaultHTTPGetAddress)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		if resp.StatusCode != 200 {
			return events.APIGatewayProxyResponse{}, ErrNon200Response
		}

		ip, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		if len(ip) == 0 {
			return events.APIGatewayProxyResponse{}, ErrNoIP
		}

		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Hello World, %v", string(ip)),
			StatusCode: 200,
		}, nil
	}

	photoName, found := request.QueryStringParameters["name"]

	if found {
		svc := dynamodb.New(sess)

		tableName := "photograph" // photoCaptureTime := "1972-12-31T15:50:09Z"

		result, err := svc.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"Name": {
					S: aws.String(photoName),
				},
			},
		})

		if err != nil {
			log.Fatalf("Got error calling GetItem: %s", err)
		}

		pmd := PhotoMetaData{}

		err1 := dynamodbattribute.UnmarshalMap(result.Item, &pmd)

		if err1 != nil {
			log.Fatalf("Got error calling GetItem: %s", err)
		}

		pmdjson, errMarshal := json.Marshal(pmd)

		if errMarshal != nil {
			fmt.Printf("error marshaling object: %v", err)
		}

		return events.APIGatewayProxyResponse{
			Body:       string(pmdjson),
			StatusCode: 200,
		}, nil
	}

	svc := s3.New(sess)

	resp, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String("photos-source-eick-com"),
	})

	if err != nil {
		log.Fatalln(err.Error())
	}

	total := 0
	for _, key := range resp.Contents {
		total++
		log.Printf("key: %v", key.Key)
	}

	log.Printf("total keys: %v", total)

	return events.APIGatewayProxyResponse{
		Body:       string(fmt.Sprintf("total keys: %v", total)),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}

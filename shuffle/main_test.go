package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	t.Run("Unable to get IP", func(t *testing.T) {
		DefaultHTTPGetAddress = "http://127.0.0.1:12345"

		_, err := handler(events.APIGatewayProxyRequest{})
		if err == nil {
			t.Fatal("Error failed to trigger with an invalid request")
		}
	})

	t.Run("Non 200 Response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer ts.Close()

		DefaultHTTPGetAddress = ts.URL

		_, err := handler(events.APIGatewayProxyRequest{})
		if err != nil && err.Error() != ErrNon200Response.Error() {
			t.Fatalf("Error failed to trigger with an invalid HTTP response: %v", err)
		}
	})

	t.Run("Unable decode IP", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer ts.Close()

		DefaultHTTPGetAddress = ts.URL

		_, err := handler(events.APIGatewayProxyRequest{})
		if err == nil {
			t.Fatal("Error failed to trigger with an invalid HTTP response")
		}
	})

	t.Run("Successful Request", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			fmt.Fprintf(w, "127.0.0.1")
		}))
		defer ts.Close()

		DefaultHTTPGetAddress = ts.URL

		_, err := handler(events.APIGatewayProxyRequest{})
		if err != nil {
			t.Fatal("Everything should be ok")
		}
	})
}

func TestPicture(t *testing.T) {
	// Exif
	var myTests = []struct {
		inputFile     string
		exifKey       string
		exifValue     string
		expectedError error
	}{
		{inputFile: "/Users/maeick/code/personal/mms/testdata/93front-a.tif",
			exifKey:       string(exif.ImageDescription),
			exifValue:     "In Summer 1930, Helen Heber and Harry",
			expectedError: nil,
		},
		{inputFile: "/Users/maeick/code/personal/mms/testdata/93front-a.tif",
			exifKey:       string(exif.DateTime),
			exifValue:     "2021:02:14 17:57:18",
			expectedError: nil,
		},
		{inputFile: "/Users/maeick/code/personal/mms/testdata/93front-a.tif",
			exifKey:       string(exif.Artist),
			exifValue:     "Emma Eick",
			expectedError: nil,
		},
		{inputFile: "/Users/maeick/code/personal/mms/testdata/becky_20_53078.tif",
			exifKey:       string(exif.ImageDescription),
			exifValue:     "",
			expectedError: nil,
		},
		{inputFile: "/Users/maeick/code/personal/mms/testdata/becky_20_53078.tif",
			exifKey:       string(exif.DateTime),
			exifValue:     "1980:12:31 16:36:47",
			expectedError: nil,
		},
		{inputFile: "/Users/maeick/code/personal/mms/testdata/becky_20_53078.tif",
			exifKey:       string(exif.Artist),
			exifValue:     "Emma Eick",
			expectedError: nil,
		},
	}

	for _, tt := range myTests {
		fileByetes, err := os.Open(tt.inputFile)
		if err != nil {
			fmt.Print(err)
		}
		defer fileByetes.Close()

		x, err := exif.Decode(fileByetes)

		if err != nil {
			t.Errorf("Should not get an error")
		}

		exifValue, actualError := x.Get(exif.FieldName(tt.exifKey))

		if actualError != tt.expectedError {
			t.Errorf("Should not get an error")
		}

		require.Equal(t, tt.expectedError, actualError, "optional message here")

		exifValueString := strings.Trim(exifValue.String(), "\"")

		if exifValueString != tt.exifValue {
			t.Errorf("Should have gotten expected output")
		}

		require.Equal(t, tt.exifValue, exifValueString)
	}

	fileByetes, err := os.Open("/Users/maeick/code/personal/mms/testdata/becky_20_53078.tif")
	if err != nil {
		fmt.Print(err)
	}
	defer fileByetes.Close()

	x, _ := exif.Decode(fileByetes)
	title, _ := x.Get(exif.UserComment)
	fmt.Printf("title is: %s\n", title)
	// photo.Title = getCleanExifValue(title)

	caption, _ := x.Get(exif.ImageDescription)
	fmt.Printf("caption is: %s\n", caption)
	// photo.Caption = getCleanExifValue(caption)

	captureTime, _ := x.DateTime()
	fmt.Printf("captureTime is: %s\n", captureTime)

	artist, _ := x.Get(exif.Artist)
	fmt.Printf("artist is: %s\n", artist)

	// photo.CaptureTime = captureTime

}

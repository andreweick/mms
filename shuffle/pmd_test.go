package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTiff(t *testing.T) {
	// Exif
	var myTests = []struct {
		inputFile     string
		Artist        string
		Description   string
		expectedError error
	}{
		{inputFile: "/Users/maeick/code/personal/mms/testdata/tiff/93front-a.tif",
			Artist:       "Emma Eick",
			Description:   "In Summer 1930, Helen Heber and Harry",
			expectedError: nil,
		},
		{inputFile: "/Users/maeick/code/personal/mms/testdata/tiff/Andy-kid-basketball.tif",
			Artist:        "Harry Eick",
			Description:   "Andy dunking basketball in snow in East Lansing, circa 1976",
			expectedError: nil,
		},
	}

	for _, tt := range myTests {

		pmd := populatePMD(tt.inputFile)

		assert.Equal(t, tt.Artist, pmd.Artist, "Filename: %v, Wanted: %v, Expected: %v", tt.inputFile, tt.Artist, pmd.Artist)
		assert.Equal(t, tt.Description, pmd.Description, "Filename: %v, Wanted: %v, Expected: %v\n", tt.inputFile, tt.Description, pmd.Description)
	}
}

func TestJpeg(t *testing.T) {
	// Exif
	var myTests = []struct {
		inputFile     string
		Artist        string
		Description   string
		expectedError error
	}{
		{inputFile: "/Users/maeick/code/personal/mms/testdata/jpeg/93front-a.jpg",
			Artist:       "Emma Eick",
			Description:   "In Summer 1930, Helen Heber and Harry",
			expectedError: nil,
		},
		{inputFile: "/Users/maeick/code/personal/mms/testdata/jpeg/Andy-kid-basketball.jpg",
			Artist:        "Harry Eick",
			Description:   "Andy dunking basketball in snow in East Lansing, circa 1976",
			expectedError: nil,
		},
	}

	for _, tt := range myTests {

		pmd := populatePMD(tt.inputFile)

		assert.Equal(t, tt.Artist, pmd.Artist, "Filename: %v, Wanted: %v, Expected: %v", tt.inputFile, tt.Artist, pmd.Artist)
		assert.Equal(t, tt.Description, pmd.Description, "Filename: %v, Wanted: %v, Expected: %v\n", tt.inputFile, tt.Description, pmd.Description)
	}
}

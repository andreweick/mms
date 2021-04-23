package main

import (
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"os"
	"time"
)

type Labels struct {
	Name       string
	Confidence float64
}

type PhotoMetaData struct {
	ID             string
	Key            string
	Name           string
	ParsedName     string
	Artist         string
	CaptureTime    time.Time
	Description    string
	Caption        string
	PerceptualHash uint64
	Classification struct {
		Labels []Labels
	}
}

func getCleanExifValue(md *tiff.Tag) string {
	if md == nil {
		return ""
	}
	s := fmt.Sprintf("%v", md)

	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}

func populatePMD(filepath string) *PhotoMetaData {
	fileBytes, err := os.Open(filepath)
	if err != nil {
		fmt.Print(err)
	}

	defer fileBytes.Close()

	x, err := exif.Decode(fileBytes)

	if err != nil {
		fmt.Errorf("should not get an error")
	}

	var pmd *PhotoMetaData
	pmd = new(PhotoMetaData)

	exifValueArtist, err := x.Get(exif.Artist)

	if err != nil {
		fmt.Errorf("error decoding the Artist")
	}

	pmd.Artist = getCleanExifValue(exifValueArtist)

	pmd.CaptureTime, err = x.DateTime()

	if err != nil {
		fmt.Errorf("error decodeing the time")
	}

	exifValueDescription, err := x.Get(exif.ImageDescription)

	if err != nil {
		fmt.Errorf("error decoding the description")
	}

	pmd.Description = getCleanExifValue(exifValueDescription)

	return pmd
}

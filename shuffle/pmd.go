package main

import (
	"fmt"
	_ "image/jpeg"
	"os"
	"strconv"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

type Labels struct {
	Name       string
	Confidence float64
}

type PhotoMetaData struct {
	Name           string
	ParsedName     string
	Artist         string
	CaptureTime    time.Time
	Description    string
	Caption        string
	ID             uint64
	Height         int
	Width          int
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
		fmt.Print("should not get an error")
	}

	var pmd *PhotoMetaData
	pmd = new(PhotoMetaData)

	exifValueArtist, err := x.Get(exif.Artist)

	if err != nil {
		fmt.Print("error decoding the Artist")
	}

	pmd.Artist = getCleanExifValue(exifValueArtist)

	pmd.CaptureTime, err = x.DateTime()

	if err != nil {
		fmt.Print("error decodeing the time")
	}

	exifValueDescription, err := x.Get(exif.ImageDescription)

	if err != nil {
		fmt.Print("error decoding the description")
	}

	pmd.Description = getCleanExifValue(exifValueDescription)

	exifVaultWidth, _ := x.Get(exif.ImageWidth)
	pmd.Width, _ = strconv.Atoi(getCleanExifValue(exifVaultWidth))

	exifVaultLength, _ := x.Get(exif.ImageLength) // Image height called
	pmd.Height, _ = strconv.Atoi(getCleanExifValue(exifVaultLength))

	return pmd
}

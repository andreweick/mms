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

	var pmd *PhotoMetaData = new(PhotoMetaData)

	exifValueArtist, err := x.Get(exif.Artist)

	if err != nil {
		fmt.Print("error decoding the Artist")
	}

	pmd.Artist = getCleanExifValue(exifValueArtist)

	pmd.CaptureTime, err = x.DateTime()

	if err != nil {
		fmt.Print("error decodeing the time")
	}

	// `Format` and `Parse` use example-based layouts. Usually
	// you'll use a constant from `time` for these layouts, but
	// you can also supply custom layouts. Layouts must use the
	// reference time `Mon Jan 2 15:04:05 MST 2006` to show the
	// pattern with which to format/parse a given time/string.
	// The example time must be exactly as shown: the year 2006,
	// 15 for the hour, Monday for the day of the week, etc.
	pmd.CaptureYear = pmd.CaptureTime.Format("2006")
	pmd.CaptureYearMonth = pmd.CaptureTime.Format("2006-01")
	pmd.CaptureYearMonthDay = pmd.CaptureTime.Format("2006-01-02")

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

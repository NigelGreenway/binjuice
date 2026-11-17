package bin

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGenerateBinDataUsesEnvVarsCorreclty(t *testing.T) {
	t.Setenv("BIN_DATA", "true")

	tests := []struct {
		numberOfBins string
		data         []BinData
	}{
		{
			"1",
			[]BinData{
				{
					Name:         "Test Name",
					FriendlyName: "Test Friendly Name",
					ImagePath:    "test_image.png",
					Freq:         1,
				},
			},
		},
		{
			"3",
			[]BinData{
				{
					Name:         "Test Name 1",
					FriendlyName: "Test Friendly Name 1",
					ImagePath:    "test_image_1.png",
					Freq:         1,
				},
				{
					Name:         "Test Name 2",
					FriendlyName: "Test Friendly Name 2",
					ImagePath:    "test_image_2.png",
					Freq:         2,
				},
				{
					Name:         "Test Name 3",
					FriendlyName: "Test Friendly Name 3",
					ImagePath:    "test_image_3.png",
					Freq:         1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Setenv("NUMBER_OF_BINS", test.numberOfBins)

		for i, testBinData := range test.data {
			envName := "BIN_" + strconv.Itoa(i+1)
			envValue := fmt.Sprintf("%s,%s,%s,%d", testBinData.Name, testBinData.FriendlyName, testBinData.ImagePath, testBinData.Freq)
			t.Setenv(envName, envValue)
		}

		data, _ := generateBinData()

		if reflect.DeepEqual(test.data, data) == false {
			t.Errorf("generateBinData() does not match expected: %v != %v", test.data, data)
		}
	}
}

func TestGenerateBinDataUsesCausesLogs(t *testing.T) {
	var stdOut bytes.Buffer

	t.Setenv("BIN_DATA", "true")
	t.Setenv("NUMBER_OF_BINS", "string")

	log.SetOutput(&stdOut)
	generateBinData()

	logMessage := stdOut.String()

	if strings.Contains(logMessage, "`NUMBER_OF_BINS` is either not set or not an integer; reverting to default setup") == false {
		t.Error("generateBinData() has not failed on on NUMBER_OF_BINS")
	}
}

func TestGenerateBinDataUsesCausesLogWhenBinItemOutOfIndexRange(t *testing.T) {
	t.Setenv("BIN_DATA", "true")
	t.Setenv("NUMBER_OF_BINS", "1")

	_, err := generateBinData()

	if err == errors.New("unable to continue due to no env vars of BIN_1 set") {
		t.Error("generateBinData() has not failed on on BIN_1")
	}
}

func TestGenerateBinDataUsesCausesLogWhenBinConfigHasIncorrectSetup(t *testing.T) {
	t.Setenv("BIN_DATA", "true")
	t.Setenv("NUMBER_OF_BINS", "1")
	t.Setenv("BIN_1", "Food,Food Waste,food_bin.png,nopes")

	_, err := generateBinData()

	if err == errors.New("unable to continue as bin frequency is not an integer") {
		t.Error("generateBinData() has not failed on on BIN_1")
	}
}

func TestNextCollectionData(t *testing.T) {
	res := NextCollectionData()
	if len(res) != 4 {
		t.Errorf("NextCollectionData() length (%v)", res)
	}
}

func TestNextCollectionDate(t *testing.T) {
	COLLECTIONDAY = time.Thursday

	tests := []struct {
		tYear  int
		tMonth time.Month
		tDay   int
		eYear  int
		eMonth time.Month
		eDay   int
	}{
		{2023, 7, 24, 2023, 7, 27},
		{2023, 7, 27, 2023, 7, 27},
		{2023, 7, 30, 2023, 8, 3},
		{2023, 9, 15, 2023, 9, 21},
		{2024, 4, 5, 2024, 4, 11},
		{2022, 2, 20, 2022, 2, 24},
	}

	for _, test := range tests {
		testDate := time.Date(test.tYear, test.tMonth, test.tDay, 1, 0, 0, 0, time.Local)
		res := NextCollectionDate(testDate)

		if res.Day() != test.eDay {
			t.Errorf("NextCollectionDate(%v) = %v", testDate, res)
		}

	}
}

func TestCurrentWeekYear(t *testing.T) {
	year, week := NextCollectionWeekYear()

	if len(strconv.Itoa(year)) != 4 {
		t.Errorf("CurrentWeekYear year length (%v)", year)
	}
	if week < 1 || week > 53 {
		t.Errorf("CurrentWeekYear week out of range (%v)", week)
	}
}

func TestWeeksInYear(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{2000, 53},
		{2004, 53},
		{2020, 53},
		{2032, 53},
		{2044, 53},
		{2001, 52},
		{2005, 52},
		{2011, 52},
		{2021, 52},
		{2023, 52},
	}

	for _, test := range tests {
		if got := WeeksInYear(test.input); got != test.want {
			t.Errorf("weeksInYear(%v) = %v", test.input, got)
		}
	}
}

func TestWeeksSinceAnchor(t *testing.T) {
	tests := []struct {
		cYear int
		cWeek int
		aYear int
		aWeek int
		want  int
	}{
		{2023, 29, 2022, 35, 46},
		{2023, 29, 2022, 34, 47},
		{2023, 29, 2022, 16, 65},
		{2025, 4, 2018, 30, 340},
	}

	for _, test := range tests {
		got, err := WeeksSinceAnchor(test.cYear, test.cWeek, test.aYear, test.aWeek)
		if err != nil {
			t.Errorf("weeksSinceAnchor(%v, %v, %v, %v) = %v", test.cYear, test.cWeek, test.aYear, test.aWeek, err)
		}

		if got != test.want {
			t.Errorf("weeksSinceAnchor(%v, %v, %v, %v) = %v", test.cYear, test.cWeek, test.aYear, test.aWeek, got)
		}
	}
}

func TestIsBinCollected(t *testing.T) {
	tests := []struct {
		bd   BinData
		ws   int
		want bool
	}{
		{BinData{Freq: 3}, 21, true},
		{BinData{Freq: 2}, 48, true},
		{BinData{Freq: 1}, 5, true},
		{BinData{Freq: 3}, 22, false},
		{BinData{Freq: 2}, 39, false},
	}

	for _, test := range tests {
		if res := IsBinCollected(test.bd, test.ws); res != test.want {
			t.Errorf("isBinCollected(%v, %v) = %v", test.bd, test.ws, test.want)
		}
	}
}

package climate

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.com/go-playground/validator/v10"
	servirtium "github.com/servirtium/servirtium-go"
	"github.com/stretchr/testify/suite"
)

type ClimateTestSuiteRecord struct {
	recordClient ClientImpl
	suite.Suite
	servirtium *servirtium.Impl
}

func TestClimateTestSuiteRecord(t *testing.T) {
	suite.Run(t, new(ClimateTestSuiteRecord))
}

func (s *ClimateTestSuiteRecord) BeforeTest(suiteName, testName string) {
	validate := validator.New()
	s.servirtium = servirtium.NewServirtium()
	s.servirtium.ReplaceRequestHeaders(map[string]string{"User-Agent": "Servirtium-Testing"})
	s.servirtium.MaskResponseHeaders(map[string]string{"Set-Cookie": "REPLACED-IN-RECORDING", "Date": "Tue, 04 Aug 2020 16:53:25 GMT"})
	passwordRegex := regexp.MustCompile(`(<password>.{0,}<\/password>)`)
	s.servirtium.MaskResponseBody(map[*regexp.Regexp]string{passwordRegex: "<password>MASKED</password>"})
	s.servirtium.StartRecord("http://climatedataapi.worldbank.org")
	recordClient := NewClient(http.DefaultClient, validate, s.servirtium.ServerRecord.URL)
	s.recordClient = *recordClient
}

func (s *ClimateTestSuiteRecord) AfterTest(suite, testName string) {
	isMatch := s.servirtium.CheckMarkdownIsDifferentToPreviousRecording(testName)
	s.True(isMatch)
	s.servirtium.WriteRecord(testName)
	s.servirtium.EndRecord()
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForGreatBritainFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForEgyptFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, recordResult)
	s.Error(recordErr)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, recordResult)
	s.Error(recordErr)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForGreatBritainAndFranceFrom1980to1999Exists() {
	var (
		ctx       = context.Background()
		expected1 = float64(988.8454972331014)
		expected2 = 913.7986955122727
	)
	record1Result, record1Err := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected1, record1Result)
	s.Nil(record1Err)
	record2Result, record2Err := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected2, record2Result)
	s.Nil(record2Err)
}

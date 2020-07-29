package climate

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
)

type ClimateRecordCompareMarkdownTestSuite struct {
	recordClient ClientImpl
	suite.Suite
	servirtium *ServirtiumImpl
}

func TestClimateRecordCompareMarkdownTestSuite(t *testing.T) {
	suite.Run(t, new(ClimateRecordCompareMarkdownTestSuite))
}

func (s *ClimateRecordCompareMarkdownTestSuite) BeforeTest(suiteName, testName string) {
	validate := validator.New()
	servirtium := NewServirtium()
	s.servirtium = servirtium
	s.servirtium.StartRecord("http://climatedataapi.worldbank.org")
	recordClient := NewClient(http.DefaultClient, validate, s.servirtium.ServerRecord.URL)
	s.recordClient = *recordClient
}

func (s *ClimateRecordCompareMarkdownTestSuite) AfterTest(suite, testName string) {
	isMatch := s.servirtium.CheckMarkdownIsDifferentToPreviousRecording(testName)
	s.True(isMatch)
	s.servirtium.EndRecord()
}

func (s *ClimateRecordCompareMarkdownTestSuite) TestAverageRainfallForGreatBritainFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateRecordCompareMarkdownTestSuite) TestAverageRainfallForFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateRecordCompareMarkdownTestSuite) TestAverageRainfallForEgyptFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateRecordCompareMarkdownTestSuite) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, recordResult)
	s.Error(recordErr)
}

func (s *ClimateRecordCompareMarkdownTestSuite) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, recordResult)
	s.Error(recordErr)
}

package climate

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
)

type ClimateRecordTestSuite struct {
	recordClient ClientImpl
	suite.Suite
	servirtium *ServirtiumImpl
}

func TestClimateRecordTestSuite(t *testing.T) {
	suite.Run(t, new(ClimateRecordTestSuite))
}

func (s *ClimateRecordTestSuite) BeforeTest(suiteName, testName string) {
	validate := validator.New()
	servirtium := NewServirtium()
	s.servirtium = servirtium
	s.servirtium.StartRecord("http://climatedataapi.worldbank.org")
	recordClient := NewClient(http.DefaultClient, validate, s.servirtium.ServerRecord.URL)
	s.recordClient = *recordClient
}

func (s *ClimateRecordTestSuite) AfterTest(suite, testName string) {
	s.servirtium.WriteRecord(testName)
	s.servirtium.EndRecord()
}

func (s *ClimateRecordTestSuite) TestAverageRainfallForGreatBritainFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateRecordTestSuite) TestAverageRainfallForFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateRecordTestSuite) TestAverageRainfallForEgyptFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateRecordTestSuite) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, recordResult)
	s.Error(recordErr)
}

func (s *ClimateRecordTestSuite) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, recordResult)
	s.Error(recordErr)
}

func (s *ClimateRecordTestSuite) TestAverageRainfallForGreatBritainAndFranceFrom1980to1999Exists() {
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

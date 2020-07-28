package climate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
)

type ClimateRecordTestSuite struct {
	recordClient ClientImpl
	suite.Suite
	serverListenWithPort *httptest.Server
}

func TestClimateRecordTestSuite(t *testing.T) {
	suite.Run(t, new(ClimateRecordTestSuite))
}

func (s *ClimateRecordTestSuite) TearDownTest() {
	s.serverListenWithPort.Close()
}

func (s *ClimateRecordTestSuite) SetupTest() {
	validate := validator.New()
	recordClient := NewClient(http.DefaultClient, validate, "http://localhost:61417")
	s.recordClient = *recordClient
}

func (s *ClimateRecordTestSuite) BeforeTest(suiteName, testName string) {
	ts := ManInTheMiddle(testName)
	s.serverListenWithPort = ts
	s.serverListenWithPort.Start()
	cleanUpMarkdownFile(testName)
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

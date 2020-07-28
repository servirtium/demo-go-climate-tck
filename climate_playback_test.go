package climate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
)

type ClimatePlaybackTestSuite struct {
	playbackClient ClientImpl
	suite.Suite
	serverPlaybackMock *httptest.Server
}

func TestClimatePlaybackTestSuite(t *testing.T) {
	suite.Run(t, new(ClimatePlaybackTestSuite))
}

func (s *ClimatePlaybackTestSuite) BeforeTest(suiteName, testName string) {
	validate := validator.New()
	s.serverPlaybackMock = serverPlaybackMock(testName)
	playbackClient := NewClient(http.DefaultClient, validate, s.serverPlaybackMock.URL)
	s.playbackClient = *playbackClient
}

func (s *ClimatePlaybackTestSuite) TestAverageRainfallForGreatBritainFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, playbackResult)
	s.Nil(playbackErr)
}

func (s *ClimatePlaybackTestSuite) TestAverageRainfallForFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, playbackResult)
	s.Nil(playbackErr)
}

func (s *ClimatePlaybackTestSuite) TestAverageRainfallForEgyptFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, playbackResult)
	s.Nil(playbackErr)
}

func (s *ClimatePlaybackTestSuite) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, playbackResult)
	s.Error(playbackErr)
}

func (s *ClimatePlaybackTestSuite) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, playbackResult)
	s.Error(playbackErr)
}

package climate

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	servirtium "github.com/servirtium/servirtium-go"
	"github.com/stretchr/testify/suite"
)

type ClimateTestSuitePlayback struct {
	playbackClient ClientImpl
	suite.Suite
	servirtium *servirtium.Impl
}

func TestClimateTestSuitePlayback(t *testing.T) {
	suite.Run(t, new(ClimateTestSuitePlayback))
}

func (s *ClimateTestSuitePlayback) BeforeTest(suiteName, testName string) {
	validate := validator.New()
	servirtium := servirtium.NewServirtium()
	s.servirtium = servirtium
	go s.servirtium.StartPlayback(testName, 61417)
	playbackClient := NewClient(http.DefaultClient, validate, s.servirtium.ServerPlayback.Addr)
	s.playbackClient = *playbackClient
}

func (s *ClimateTestSuitePlayback) AfterTest(suite, testName string) {
	s.servirtium.EndPlayback()
	time.Sleep(4 * time.Second)
}

func (s *ClimateTestSuitePlayback) TestAverageRainfallForGreatBritainFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, playbackResult)
	s.Nil(playbackErr)
}

func (s *ClimateTestSuitePlayback) TestAverageRainfallForFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, playbackResult)
	s.Nil(playbackErr)
}

func (s *ClimateTestSuitePlayback) TestAverageRainfallForEgyptFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, playbackResult)
	s.Nil(playbackErr)
}

func (s *ClimateTestSuitePlayback) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, playbackResult)
	s.Error(playbackErr)
}

func (s *ClimateTestSuitePlayback) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, playbackResult)
	s.Error(playbackErr)
}

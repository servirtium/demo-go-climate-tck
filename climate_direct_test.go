package climate

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
)

type ClimateTestSuiteDirect struct {
	directClient ClientImpl
	suite.Suite
}

func TestClimateTestSuiteDirect(t *testing.T) {
	suite.Run(t, new(ClimateTestSuiteDirect))
}

func (s *ClimateTestSuiteDirect) SetupTest() {
	validate := validator.New()
	directClient := NewClient(http.DefaultClient, validate, "https://servirtium.github.io/worldbank-climate-recordings")
	s.directClient = *directClient
}

func (s *ClimateTestSuiteDirect) TestAverageRainfallForGreatBritainFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	directResult, directErr := s.directClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, directResult)
	s.Nil(directErr)
}

func (s *ClimateTestSuiteDirect) TestAverageRainfallForFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	directResult, directErr := s.directClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, directResult)
	s.Nil(directErr)
}

func (s *ClimateTestSuiteDirect) TestAverageRainfallForEgyptFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	directResult, directErr := s.directClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, directResult)
	s.Nil(directErr)
}

func (s *ClimateTestSuiteDirect) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(-1)
	)
	directResult, directErr := s.directClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, directResult)
	s.Error(directErr)
}

func (s *ClimateTestSuiteDirect) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(-1)
	)
	directResult, directErr := s.directClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, directResult)
	s.Error(directErr)
}

func (s *ClimateTestSuiteDirect) TestAverageRainfallForGreatBritainAndFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(1902.644192745374)
	)
	recordResult, recordErr := s.directClient.GetAveAnnualRainfallMany(ctx, 1980, 1999, "gbr", "fra")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

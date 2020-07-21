package climate

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
)

type ClimateTestSuite struct {
	suite.Suite
	client ClientImpl
}

func TestClimateTestSuite(t *testing.T) {
	suite.Run(t, new(ClimateTestSuite))
}

func (s *ClimateTestSuite) SetupTest() {
	validate := validator.New()
	client := NewClient(http.DefaultClient, validate)
	s.client = *client
}

func (s *ClimateTestSuite) TestNewClient_Success() {
	s.NotNil(s.client)
}

func (s *ClimateTestSuite) TestNewGetRequestWithRelativeURL_Success() {
	var (
		input    = "/country/annualavg/pr/1980/1999/GBR.xml"
		expected = "http://climatedataapi.worldbank.org/climateweb/rest/v1/country/annualavg/pr/1980/1999/GBR.xml"
		ctx      = context.Background()
	)
	r, err := s.client.NewGetRequest(ctx, input)
	s.Equal(expected, r.URL.String())
	s.Nil(err)
}

func (s *ClimateTestSuite) TestNewGetRequestWithAbsoluteURL_Success() {
	var (
		input    = "http://climatedataapi.worldbank.org/climateweb/rest/v1/country/annualavg/pr/1980/1999/GBR.xml"
		expected = "http://climatedataapi.worldbank.org/climateweb/rest/v1/country/annualavg/pr/1980/1999/GBR.xml"
		ctx      = context.Background()
	)
	r, err := s.client.NewGetRequest(ctx, input)
	s.Equal(expected, r.URL.String())
	s.Nil(err)
}

func (s *ClimateTestSuite) TestGetAnnualRainfall_Success() {
	var (
		input = GetAnnualRainfallArgs{
			FromCCYY:   "1980",
			ToCCYY:     "1999",
			CountryISO: "GBR",
		}
		ctx = context.Background()
	)
	result, err := s.client.GetAnnualRainfall(ctx, input)
	s.NotNil(result)
	s.Nil(err)
}

func (s *ClimateTestSuite) TestGetAnnualRainfall_Failed() {
	var (
		input = GetAnnualRainfallArgs{
			FromCCYY:   "1980",
			ToCCYY:     "1999",
			CountryISO: "GB",
		}
		ctx      = context.Background()
		expected = List{}
	)
	result, err := s.client.GetAnnualRainfall(ctx, input)
	s.Equal(expected, result)
	s.NotNil(err)
}

func (s *ClimateTestSuite) TestAverageRainfallForGreatBritainFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	result, err := s.client.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, result)
	s.Nil(err)
}

func (s *ClimateTestSuite) TestAverageRainfallForFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	result, err := s.client.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, result)
	s.Nil(err)
}

func (s *ClimateTestSuite) TestAverageRainfallForEgyptFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	result, err := s.client.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, result)
	s.Nil(err)
}

func (s *ClimateTestSuite) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	result, err := s.client.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, result)
	s.Error(err)
}

func (s *ClimateTestSuite) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	result, err := s.client.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, result)
	s.Error(err)
}

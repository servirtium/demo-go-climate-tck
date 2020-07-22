package climate

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

func serverMock() *httptest.Server {
	r := mux.NewRouter()
	r.HandleFunc("/country/annualavg/pr/{fromCCYY}/{toCCYY}/{countryISO}.xml", anualAvgHandler)
	srv := httptest.NewServer(r)
	return srv
}

func anualAvgHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fromCCYY := vars["fromCCYY"]
	toCCYY := vars["toCCYY"]
	countryISO := vars["countryISO"]

	data, err := ioutil.ReadFile(fmt.Sprintf("./mock/%s_%s_%s.xml", fromCCYY, toCCYY, countryISO))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/xml")
	_, _ = w.Write(data)
}

type ClimateTestSuite struct {
	suite.Suite
	client     ClientImpl
	serverMock *httptest.Server
}

func TestClimateTestSuite(t *testing.T) {
	suite.Run(t, new(ClimateTestSuite))
}

func (s *ClimateTestSuite) SetupTest() {
	validate := validator.New()
	s.serverMock = serverMock()
	// Flag to ON/OFF for request real API and mock server API
	// s.serverMock.URL = "http://climatedataapi.worldbank.org/climateweb/rest/v1"
	client := NewClient(http.DefaultClient, validate, s.serverMock.URL)
	s.client = *client
}

func (s *ClimateTestSuite) TestNewClient_Success() {
	s.NotNil(s.client)
}

func (s *ClimateTestSuite) TestNewGetRequestWithRelativeURL_Success() {
	var (
		input    = "/country/annualavg/pr/1980/1999/GBR.xml"
		expected = fmt.Sprintf("%s/country/annualavg/pr/1980/1999/GBR.xml", s.serverMock.URL)
		ctx      = context.Background()
	)
	r, err := s.client.NewGetRequest(ctx, input)
	s.Equal(expected, r.URL.String())
	s.Nil(err)
}

func (s *ClimateTestSuite) TestNewGetRequestWithAbsoluteURL_Success() {
	var (
		input    = fmt.Sprintf("%s/country/annualavg/pr/1980/1999/GBR.xml", s.serverMock.URL)
		expected = fmt.Sprintf("%s/country/annualavg/pr/1980/1999/GBR.xml", s.serverMock.URL)
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

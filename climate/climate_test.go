package climate

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

func serverMock() *httptest.Server {
	r := mux.NewRouter()
	r.HandleFunc("/country/annualavg/pr/{fromCCYY}/{toCCYY}/{countryISO}.xml", anualAvgHandlerRemote)
	srv := httptest.NewServer(r)
	return srv
}

func anualAvgHandlerLocal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fromCCYY := vars["fromCCYY"]
	toCCYY := vars["toCCYY"]
	countryISO := vars["countryISO"]

	data, err := ioutil.ReadFile(fmt.Sprintf("./mock/%s_%s_%s.xml", fromCCYY, toCCYY, countryISO))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/xml")
	_, _ = w.Write(data)
	return
}

func anualAvgHandlerRemote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fromCCYY := vars["fromCCYY"]
	toCCYY := vars["toCCYY"]
	countryISO := vars["countryISO"]
	validate := validator.New()
	client := NewClient(&http.Client{}, validate, "http://climatedataapi.worldbank.org/climateweb/rest/v1")
	list, err := client.GetAnnualRainfall(r.Context(), GetAnnualRainfallArgs{
		FromCCYY:   fromCCYY,
		ToCCYY:     toCCYY,
		CountryISO: countryISO,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
		return
	}
	output, err := xml.MarshalIndent(list, "  ", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/xml")
	_, _ = w.Write([]byte(output))
	return
}

type ClimateTestSuite struct {
	client ClientImpl
	suite.Suite
	remoteClient ClientImpl
	serverMock   *httptest.Server
}

func TestClimateTestSuite(t *testing.T) {
	suite.Run(t, new(ClimateTestSuite))
}

func (s *ClimateTestSuite) SetupTest() {
	validate := validator.New()
	s.serverMock = serverMock()
	client := NewClient(http.DefaultClient, validate, s.serverMock.URL)
	s.client = *client
	remoteClient := NewClient(http.DefaultClient, validate, "http://climatedataapi.worldbank.org/climateweb/rest/v1")
	s.remoteClient = *remoteClient
}

func (s *ClimateTestSuite) TestNewClient_Success() {
	s.NotNil(s.client)
}

func (s *ClimateTestSuite) TestNewGetRequestWithRelativeURL_Success() {
	var (
		input          = "/country/annualavg/pr/1980/1999/GBR.xml"
		expected       = fmt.Sprintf("%s/country/annualavg/pr/1980/1999/GBR.xml", s.serverMock.URL)
		expectedRemote = fmt.Sprintf("%s/country/annualavg/pr/1980/1999/GBR.xml", "http://climatedataapi.worldbank.org/climateweb/rest/v1")
		ctx            = context.Background()
	)
	r, err := s.client.NewGetRequest(ctx, input)
	s.Equal(expected, r.URL.String())
	s.Nil(err)
	r, err = s.remoteClient.NewGetRequest(ctx, input)
	s.Equal(expectedRemote, r.URL.String())
	s.Nil(err)
}

func (s *ClimateTestSuite) TestNewGetRequestWithAbsoluteURL_Success() {
	var (
		input          = fmt.Sprintf("%s/country/annualavg/pr/1980/1999/GBR.xml", s.serverMock.URL)
		expected       = fmt.Sprintf("%s/country/annualavg/pr/1980/1999/GBR.xml", s.serverMock.URL)
		remoteInput    = fmt.Sprintf("%s/country/annualavg/pr/1980/1999/GBR.xml", "http://climatedataapi.worldbank.org/climateweb/rest/v1")
		expectedRemote = fmt.Sprintf("%s/country/annualavg/pr/1980/1999/GBR.xml", "http://climatedataapi.worldbank.org/climateweb/rest/v1")
		ctx            = context.Background()
	)
	r, err := s.client.NewGetRequest(ctx, input)
	s.Equal(expected, r.URL.String())
	s.Nil(err)
	remoteRq, err := s.remoteClient.NewGetRequest(ctx, remoteInput)
	s.Equal(expectedRemote, remoteRq.URL.String())
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
	result, err = s.remoteClient.GetAnnualRainfall(ctx, input)
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
	result, err = s.remoteClient.GetAnnualRainfall(ctx, input)
	s.Equal(expected, result)
	s.NotNil(err)
}

func (s *ClimateTestSuite) TestCalculateAveAnual_Success() {
	var (
		list = List{
			DomainWebAnnualGcmDatum: []DomainWebAnnualGcmDatum{
				{
					AnnualData: AnnualData{
						Double: "10",
					},
				},
				{
					AnnualData: AnnualData{
						Double: "11",
					},
				},
			},
		}
		fromCCYY = int64(1980)
		toCCYY   = int64(1990)
		expected = decimal.NewFromFloat32(10.5)
	)
	result, err := s.client.calculateAveAnual(list, fromCCYY, toCCYY)
	s.Equal(expected.String(), result.String())
	s.Nil(err)
	result, err = s.remoteClient.calculateAveAnual(list, fromCCYY, toCCYY)
	s.Equal(expected.String(), result.String())
	s.Nil(err)
}

func (s *ClimateTestSuite) TestCalculateAveAnual_Failed() {
	var (
		list = List{
			DomainWebAnnualGcmDatum: []DomainWebAnnualGcmDatum{},
		}
		fromCCYY = int64(1980)
		toCCYY   = int64(1990)
		expected = decimal.NewFromInt(0)
	)
	result, err := s.client.calculateAveAnual(list, fromCCYY, toCCYY)
	s.Equal(result, expected)
	s.NotNil(err)
	result, err = s.remoteClient.calculateAveAnual(list, fromCCYY, toCCYY)
	s.Equal(result, expected)
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
	result, err = s.remoteClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
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
	result, err = s.remoteClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
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
	result, err = s.remoteClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
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
	result, err = s.remoteClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
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
	result, err = s.remoteClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, result)
	s.Error(err)
}

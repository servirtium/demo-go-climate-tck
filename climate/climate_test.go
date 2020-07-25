package climate

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

func serverMock() *httptest.Server {
	r := mux.NewRouter()
	r.HandleFunc("/climateweb/rest/v1/country/annualavg/pr/{fromCCYY}/{toCCYY}/{countryISO}.xml", anualAvgHandlerRemote)
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

func generateFileName(requestURLPath string) string {
	params := strings.Split(requestURLPath, "/")
	countryIndex := len(params) - 1
	toCCYYIndex := len(params) - 2
	fromCCYYIndex := len(params) - 3

	countryISO := strings.ReplaceAll(params[countryIndex], ".xml", "")
	fromCCYY := params[fromCCYYIndex]
	toCCYY := params[toCCYYIndex]
	fileName := fmt.Sprintf("./record/average_Rainfall_For_%s_From_%s_to_%s.md", countryISO, fromCCYY, toCCYY)
	return fileName

}

func record(params recordData) {
	content, err := ioutil.ReadFile("./template.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.New("template").Parse(string(content))
	if err != nil {
		log.Fatal(err)
	}
	data := recordData{
		RequestMethod:       params.RequestMethod,
		RequestURLPath:      params.RequestURLPath,
		RequestHeader:       params.RequestHeader,
		RequestBody:         params.RequestBody,
		ResponseHeader:      params.ResponseHeader,
		ResponseBody:        params.ResponseBody,
		ResponseContentType: params.ResponseContentType,
		ResponseStatus:      params.ResponseStatus,
	}
	fileName := generateFileName(params.RequestURLPath)
	buffer := new(bytes.Buffer)
	tmpl.Execute(buffer, data)
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	file, err = os.Create(fileName)
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}

func anualAvgHandlerRemote(w http.ResponseWriter, r *http.Request) {
	// Clone Request Body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
	url := fmt.Sprintf("%s://%s%s", "http", "climatedataapi.worldbank.org", r.RequestURI)
	proxyReq, err := http.NewRequest(r.Method, url, bytes.NewReader(reqBody))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = r.Header
	proxyReq.Header = make(http.Header)
	for h, val := range r.Header {
		proxyReq.Header[h] = val
	}

	resp, err := http.DefaultClient.Do(proxyReq)

	// Clone resp
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newRespHeader := resp.Header
	newRespHeader.Del("Set-Cookie")
	newRespHeader.Del("Date")
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))

	record(recordData{
		RequestURLPath:      r.URL.Path,
		RequestMethod:       r.Method,
		RequestHeader:       r.Header,
		RequestBody:         string(reqBody),
		ResponseHeader:      newRespHeader,
		ResponseBody:        string(respBody),
		ResponseContentType: resp.Header.Get("Content-Type"),
		ResponseStatus:      resp.Status,
	})

	defer func() {
		_ = resp.Body.Close()
	}()
	w.Write(respBody)
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
	remoteClient := NewClient(http.DefaultClient, validate, "http://climatedataapi.worldbank.org")
	s.remoteClient = *remoteClient
}

func (s *ClimateTestSuite) TestNewClient_Success() {
	s.NotNil(s.client)
}

func (s *ClimateTestSuite) TestNewGetRequestWithRelativeURL_Success() {
	var (
		input          = "/climateweb/rest/v1/country/annualavg/pr/1980/1999/GBR.xml"
		expected       = fmt.Sprintf("%s/climateweb/rest/v1/country/annualavg/pr/1980/1999/GBR.xml", s.serverMock.URL)
		expectedRemote = fmt.Sprintf("%s/climateweb/rest/v1/country/annualavg/pr/1980/1999/GBR.xml", "http://climatedataapi.worldbank.org")
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
		input          = fmt.Sprintf("%s/climateweb/rest/v1/country/annualavg/pr/1980/1999/GBR.xml", s.serverMock.URL)
		expected       = fmt.Sprintf("%s/climateweb/rest/v1/country/annualavg/pr/1980/1999/GBR.xml", s.serverMock.URL)
		remoteInput    = fmt.Sprintf("%s/climateweb/rest/v1/country/annualavg/pr/1980/1999/GBR.xml", "http://climatedataapi.worldbank.org")
		expectedRemote = fmt.Sprintf("%s/climateweb/rest/v1/country/annualavg/pr/1980/1999/GBR.xml", "http://climatedataapi.worldbank.org")
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

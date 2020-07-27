package climate

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/suite"
)

func serverPlaybackMock(recordFileName string) *httptest.Server {
	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(anualAvgHandlerPlayback(recordFileName))
	srv := httptest.NewServer(r)
	return srv
}

func anualAvgHandlerPlayback(recordFileName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadFile(fmt.Sprintf("./mock/%s.md", recordFileName))
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
}

func ManInTheMiddle(recordFileName string) *httptest.Server {
	l, err := net.Listen("tcp", "127.0.0.1:61417")
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(manInTheMiddleHandler(recordFileName))
	ts := httptest.NewUnstartedServer(r)

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	ts.Listener.Close()
	ts.Listener = l
	return ts
}

func manInTheMiddleHandler(recordFileName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
		proxyReq.Header.Set("User-Agent", "Servirtium-Testing")

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
			RecordFileName:      recordFileName,
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
	buffer := new(bytes.Buffer)
	tmpl.Execute(buffer, data)
	file, err := os.Create(fmt.Sprintf("./mock/%s.md", params.RecordFileName))
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}

type ClimateTestSuite struct {
	recordClient   ClientImpl
	directClient   ClientImpl
	playbackClient ClientImpl
	suite.Suite
	serverPlaybackMock   *httptest.Server
	serverListenWithPort *httptest.Server
}

func TestClimateTestSuite(t *testing.T) {
	suite.Run(t, new(ClimateTestSuite))
}

func (s *ClimateTestSuite) SetupTest() {
	validate := validator.New()
	directClient := NewClient(http.DefaultClient, validate, "http://climatedataapi.worldbank.org")
	s.directClient = *directClient
	recordClient := NewClient(http.DefaultClient, validate, "http://localhost:61417")
	s.recordClient = *recordClient
}

func (s *ClimateTestSuite) TearDownTest() {
	s.serverListenWithPort.Close()
}

func (s *ClimateTestSuite) BeforeTest(suiteName, testName string) {
	validate := validator.New()
	prefixTestName := strings.Split(testName, "_")[0]
	s.serverPlaybackMock = serverPlaybackMock(prefixTestName)
	playbackClient := NewClient(http.DefaultClient, validate, s.serverPlaybackMock.URL)
	s.playbackClient = *playbackClient
	ts := ManInTheMiddle(prefixTestName)
	s.serverListenWithPort = ts
	s.serverListenWithPort.Start()
}

func (s *ClimateTestSuite) AfterTest(suiteName, testName string) {
	pretty.Println(suiteName, testName)
	// TODO: will finish the record here
}

func (s *ClimateTestSuite) TestNewClient_Success() {
	s.NotNil(s.directClient)
	s.NotNil(s.recordClient)
	s.NotNil(s.playbackClient)
}

func (s *ClimateTestSuite) TestAverageRainfallForGreatBritainFrom1980to1999Exists_DirectMode() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	directResult, directErr := s.directClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, directResult)
	s.Nil(directErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForGreatBritainFrom1980to1999Exists_RecordMode() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForGreatBritainFrom1980to1999Exists_PlaybackMode() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, playbackResult)
	s.Nil(playbackErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForGreatBritainFrom1980to1999Exists_FailIfMarkdownIsDifferentToPreviousRecording() {
	var (
		ctx = context.Background()
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Nil(playbackErr)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Nil(recordErr)
	s.Equal(playbackResult, recordResult)
}

func (s *ClimateTestSuite) TestAverageRainfallForFranceFrom1980to1999Exists_DirectMode() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	directResult, directErr := s.directClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, directResult)
	s.Nil(directErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForFranceFrom1980to1999Exists_RecordMode() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForFranceFrom1980to1999Exists_PlaybackMode() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, playbackResult)
	s.Nil(playbackErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForFranceFrom1980to1999Exists_FailIfMarkdownIsDifferentToPreviousRecording() {
	var (
		ctx = context.Background()
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Nil(playbackErr)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Nil(recordErr)
	s.Equal(playbackResult, recordResult)
}

func (s *ClimateTestSuite) TestAverageRainfallForEgyptFrom1980to1999Exists_DirectMode() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	directResult, directErr := s.directClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, directResult)
	s.Nil(directErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForEgyptFrom1980to1999Exists_RecordMode() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForEgyptFrom1980to1999Exists_PlaybackMode() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, playbackResult)
	s.Nil(playbackErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForEgyptFrom1980to1999Exists_FailIfMarkdownIsDifferentToPreviousRecording() {
	var (
		ctx = context.Background()
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Nil(recordErr)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Nil(playbackErr)
	s.Equal(playbackResult, recordResult)
}

func (s *ClimateTestSuite) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist_DirectMode() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	directResult, directErr := s.directClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, directResult)
	s.Error(directErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist_RecordMode() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, recordResult)
	s.Error(recordErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist_PlaybackMode() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, playbackResult)
	s.Error(playbackErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist_FailIfMarkdownIsDifferentToPreviousRecording() {
	var (
		ctx = context.Background()
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Error(playbackErr)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Error(recordErr)
	s.Equal(playbackResult, recordResult)
}

func (s *ClimateTestSuite) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist_DirectMode() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	directResult, directErr := s.directClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, directResult)
	s.Error(directErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist_RecordMode() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, recordResult)
	s.Error(recordErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist_PlaybackMode() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, playbackResult)
	s.Error(playbackErr)
}

func (s *ClimateTestSuite) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist_FailIfMarkdownIsDifferentToPreviousRecording() {
	var (
		ctx = context.Background()
	)
	playbackResult, playbackErr := s.playbackClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Error(playbackErr)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Error(recordErr)
	s.Equal(playbackResult, recordResult)
}

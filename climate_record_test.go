package climate

import (
	"context"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	servirtium "github.com/servirtium/servirtium-go"
	"github.com/stretchr/testify/suite"
)

type ClimateTestSuiteRecord struct {
	recordClient ClientImpl
	suite.Suite
	servirtium *servirtium.Impl
}

func TestClimateTestSuiteRecord(t *testing.T) {
	suite.Run(t, new(ClimateTestSuiteRecord))
}

func (s *ClimateTestSuiteRecord) BeforeTest(suiteName, testName string) {
	validate := validator.New()
	s.servirtium = servirtium.NewServirtium()
	s.servirtium.SetCallerRequestHeaderReplacements(map[*regexp.Regexp]string{
		regexp.MustCompile("User-Agent: (.*)"): "User-Agent: Servirtium-Agent",
	})
	s.servirtium.SetRecordResponseHeaderReplacements(map[*regexp.Regexp]string{
		regexp.MustCompile("Set-Cookie: (.*)"): "REPLACED-IN-RECORDING",
		regexp.MustCompile("Date: (.*)"):       "Date: Tue, 04 Aug 2020 16:53:25 GMT",
	})
	go s.servirtium.StartRecord("http://worldbank-api-for-servirtium.local.gd:4567", 61417)
	recordClient := NewClient(http.DefaultClient, validate, s.servirtium.ServerRecord.Addr)
	s.recordClient = *recordClient
}

func (s *ClimateTestSuiteRecord) AfterTest(suite, testName string) {
	s.servirtium.WriteRecord(testName)
	s.servirtium.EndRecord()
	time.Sleep(4 * time.Second)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForGreatBritainFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForEgyptFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Equal(expected, recordResult)
	s.Nil(recordErr)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Equal(expected, recordResult)
	s.Error(recordErr)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(0)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Equal(expected, recordResult)
	s.Error(recordErr)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForGreatBritainAndFranceFrom1980to1999Exists() {
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

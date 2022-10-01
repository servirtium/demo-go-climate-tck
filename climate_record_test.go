package climate

import (
	"context"
	"github.com/go-playground/validator/v10"
	servirtium "github.com/servirtium/servirtium-go"
	"github.com/stretchr/testify/suite"
	"net/http"
	"regexp"
	"testing"
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
	s.servirtium.SetRecordResponseHeadersRemoval([]string{"For_testing"})
	s.servirtium.SetRecordResponseHeaderReplacements(map[*regexp.Regexp]string{
		regexp.MustCompile("Set-Cookie: (.*)"): "REPLACED-IN-RECORDING",
		regexp.MustCompile("Date: (.*)"):       "Date: Tue, 04 Aug 2020 16:53:25 GMT",
	})
	s.servirtium.StartRecord("https://servirtium.github.io/worldbank-climate-recordings", 61417)
	recordClient := NewClient(http.DefaultClient, validate, s.servirtium.GetRecordURL())
	s.recordClient = *recordClient
}

func (s *ClimateTestSuiteRecord) AfterTest(suite, testName string) {
	s.servirtium.WriteRecord(testName)
	s.servirtium.EndRecord()
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForGreatBritainFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(988.8454972331014)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "gbr")
	s.Nil(recordErr)
	s.Equal(expected, recordResult)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = 913.7986955122727
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "fra")
	s.Nil(recordErr)
	s.Equal(expected, recordResult)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForEgyptFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(54.58587712129825)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "egy")
	s.Nil(recordErr)
	s.Equal(expected, recordResult)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForGreatBritainFrom1985to1995DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(-1)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1985, 1995, "gbr")
	s.Error(recordErr)
	s.Equal(expected, recordResult)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForMiddleEarthFrom1980to1999DoesNotExist() {
	var (
		ctx      = context.Background()
		expected = float64(-1)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfall(ctx, 1980, 1999, "mde")
	s.Error(recordErr)
	s.Equal(expected, recordResult)
}

func (s *ClimateTestSuiteRecord) TestAverageRainfallForGreatBritainAndFranceFrom1980to1999Exists() {
	var (
		ctx      = context.Background()
		expected = float64(1902.644192745374)
	)
	recordResult, recordErr := s.recordClient.GetAveAnnualRainfallMany(ctx, 1980, 1999, "gbr", "fra")
	s.Nil(recordErr)
	s.Equal(expected, recordResult)
}

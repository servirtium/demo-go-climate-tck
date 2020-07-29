package climate

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// IClient interface
type IClient interface {
	NewGetRequest(ctx context.Context, url string) (*http.Request, error)
	Do(r *http.Request, v interface{}) (*http.Response, error)
	GetAnnualRainfall(ctx context.Context, args GetAnnualRainfallArgs) (List, error)
	GetAveAnnualRainfall(ctx context.Context, fromCCYY, toCCYY, countryISO string) (float64, error)
}

// ClientImpl struct
type ClientImpl struct {
	baseURL  string
	validate *validator.Validate
	http     *http.Client
}

// NewClient ...
func NewClient(c *http.Client, validate *validator.Validate, baseURL string) *ClientImpl {
	if c == nil {
		c = http.DefaultClient
	}

	return &ClientImpl{
		http:     c,
		validate: validate,
		baseURL:  baseURL,
	}
}

// NewGetRequest ...
func (c *ClientImpl) NewGetRequest(ctx context.Context, url string) (*http.Request, error) {
	if len(url) == 0 {
		return nil, errors.New("invalid empty-string url")
	}

	// Assume the user has given a relative path.
	isRelativePath := url[0] == '/'
	if isRelativePath {
		url = c.baseURL + url
	}
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return r.WithContext(ctx), nil
}

// Do the request.
func (c *ClientImpl) Do(r *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	if v != nil {
		if err = xml.NewDecoder(resp.Body).Decode(v); err != nil {
			return nil, fmt.Errorf("unable to parse XML [%s %s]: %v", r.Method, r.URL.RequestURI(), err)
		}
	}
	return resp, nil
}

// AnnualData ...
type AnnualData struct {
	Text   string `xml:",chardata"`
	Double string `xml:"double"`
}

// DomainWebAnnualGcmDatum ...
type DomainWebAnnualGcmDatum struct {
	Text       string     `xml:",chardata"`
	Gcm        string     `xml:"gcm"`
	Variable   string     `xml:"variable"`
	FromYear   string     `xml:"fromYear"`
	ToYear     string     `xml:"toYear"`
	AnnualData AnnualData `xml:"annualData"`
}

// List ...
type List struct {
	XMLName                 xml.Name                  `xml:"list"`
	Text                    string                    `xml:",chardata"`
	DomainWebAnnualGcmDatum []DomainWebAnnualGcmDatum `xml:"domain.web.AnnualGcmDatum"`
}

// GetAnnualRainfallArgs ...
type GetAnnualRainfallArgs struct {
	FromCCYY   string `validate:"required,len=4"`
	ToCCYY     string `validate:"required,len=4"`
	CountryISO string `validate:"required,len=3"`
}

// GetAnnualRainfall ...
func (c *ClientImpl) GetAnnualRainfall(ctx context.Context, args GetAnnualRainfallArgs) (List, error) {
	err := c.validate.Struct(args)
	if err != nil {
		return List{}, err
	}
	apiURL := fmt.Sprintf("/climateweb/rest/v1/country/annualavg/pr/%s/%s/%s.xml", args.FromCCYY, args.ToCCYY, strings.ToLower(args.CountryISO))
	r, err := c.NewGetRequest(ctx, apiURL)
	if err != nil {
		return List{}, err
	}
	list := List{}
	if _, err = c.Do(r, &list); err != nil {
		return List{}, err
	}
	return list, nil
}

func (c *ClientImpl) calculateAveAnual(list List, fromCCYY, toCCYY int64) (decimal.Decimal, error) {
	domainWebAnnualGcmDatum := list.DomainWebAnnualGcmDatum
	totalAnualData := decimal.NewFromInt(0)
	totalDatum := int64(len(domainWebAnnualGcmDatum))
	if totalDatum < 1 {
		return decimal.NewFromInt(0), fmt.Errorf("date range %d-%d not supported", fromCCYY, toCCYY)
	}
	totalDatumDec := decimal.NewFromInt(totalDatum)
	for _, v := range domainWebAnnualGcmDatum {
		anualData, err := decimal.NewFromString(v.AnnualData.Double)
		if err != nil {
			continue
		}
		totalAnualData = totalAnualData.Add(anualData)
	}
	anualAve := totalAnualData.Div(totalDatumDec)
	return anualAve, nil
}

// GetAveAnnualRainfall ...
func (c *ClientImpl) GetAveAnnualRainfall(ctx context.Context, fromCCYY int64, toCCYY int64, countryISO string) (float64, error) {
	args := GetAnnualRainfallArgs{
		FromCCYY:   fmt.Sprintf("%d", fromCCYY),
		ToCCYY:     fmt.Sprintf("%d", toCCYY),
		CountryISO: countryISO,
	}
	list, err := c.GetAnnualRainfall(ctx, args)
	if err != nil {
		return 0, err
	}
	anualAve, err := c.calculateAveAnual(list, fromCCYY, toCCYY)
	if err != nil {
		return 0, err
	}
	result, _ := anualAve.Float64()
	return result, nil
}

// IServirtium ...
type IServirtium interface {
	StartRecord()
	WriteRecord(recordFileName string)
	CheckMarkdownIsDifferentToPreviousRecording(recordFileName string) bool
	EndRecord()
	StartPlayback(recordFileName string)
	EndPlayback(recordFileName string)
}

// ServirtiumImpl ...
type ServirtiumImpl struct {
	ServerPlayback  *httptest.Server
	ServerRecord    *httptest.Server
	RequestSequence int64
	Content         string
}

// NewServirtium ...
func NewServirtium() *ServirtiumImpl {
	return &ServirtiumImpl{
		RequestSequence: 0,
		Content:         "",
	}
}

// StartPlayback ...
func (s *ServirtiumImpl) StartPlayback(recordFileName string) {
	s.initServerPlayback(recordFileName)
}

// EndPlayback ...
func (s *ServirtiumImpl) EndPlayback(recordFileName string) {
	s.ServerPlayback.Close()
}

func (s *ServirtiumImpl) initServerPlayback(recordFileName string) {
	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(s.anualAvgHandlerPlayback(recordFileName))
	srv := httptest.NewServer(r)
	s.ServerPlayback = srv
}

func (s *ServirtiumImpl) anualAvgHandlerPlayback(recordFileName string) func(w http.ResponseWriter, r *http.Request) {
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

// StartRecord ...
func (s *ServirtiumImpl) StartRecord() {
	s.initRecordServer()
	s.ServerRecord.Start()
}

// WriteRecord ...
func (s *ServirtiumImpl) WriteRecord(recordFileName string) {
	filePath := fmt.Sprintf("./mock/%s.md", recordFileName)
	markdownExists := s.checkMarkdownExists(filePath)
	if !markdownExists {
		os.Create(filePath)
	}
	err := ioutil.WriteFile(filePath, []byte(s.Content), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

// EndRecord ...
func (s *ServirtiumImpl) EndRecord() {
	s.ServerRecord.Close()
}

func (s *ServirtiumImpl) initRecordServer() {
	l, err := net.Listen("tcp", "127.0.0.1:61417")
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(s.manInTheMiddleHandler())
	ts := httptest.NewUnstartedServer(r)

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	ts.Listener.Close()
	ts.Listener = l
	s.ServerRecord = ts
}

type recordData struct {
	RecordSequence      int64
	RequestMethod       string
	RequestURLPath      string
	RequestHeader       map[string][]string
	RequestBody         string
	ResponseHeader      map[string][]string
	ResponseBody        string
	ResponseStatus      string
	ResponseContentType string
}

func (s *ServirtiumImpl) manInTheMiddleHandler() func(w http.ResponseWriter, r *http.Request) {
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
		s.record(recordData{
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

// checkMarkdownExists ...
func (s *ServirtiumImpl) checkMarkdownExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (s *ServirtiumImpl) appendContentInFile(currentContent, newContent string) string {
	if s.RequestSequence == 0 {
		return newContent
	}
	finalContent := fmt.Sprintf("%s\n%s", currentContent, newContent)
	return finalContent
}

func (s *ServirtiumImpl) record(params recordData) {
	content, err := ioutil.ReadFile("./template.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.New("template").Parse(string(content))
	if err != nil {
		log.Fatal(err)
	}
	data := recordData{
		RecordSequence:      s.RequestSequence,
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
	newContent := buffer.Bytes()
	finalContent := s.appendContentInFile(s.Content, string(newContent))
	s.Content = finalContent
	s.RequestSequence = s.RequestSequence + 1
}

// CheckMarkdownIsDifferentToPreviousRecording ...
func (s *ServirtiumImpl) CheckMarkdownIsDifferentToPreviousRecording(recordFileName string) bool {
	filePath := fmt.Sprintf("./mock/%s.md", recordFileName)
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	return s.Content == string(fileContent)
}

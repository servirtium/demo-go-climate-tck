package climate

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/go-playground/validator/v10"
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

type recordData struct {
	RequestMethod       string
	RequestURLPath      string
	RequestHeader       map[string][]string
	RequestBody         string
	ResponseHeader      map[string][]string
	ResponseBody        string
	ResponseStatus      string
	ResponseContentType string
}

func (c *ClientImpl) generateFileName(requestURLPath string) string {
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

func (c *ClientImpl) record(params recordData) {
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
	fileName := c.generateFileName(params.RequestURLPath)
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

// Do the request.
func (c *ClientImpl) Do(r *http.Request, v interface{}) (*http.Response, error) {
	// Clone Request Body
	var reqBodyBytes []byte
	if r.Body != nil {
		reqBodyBytes, _ = ioutil.ReadAll(r.Body)
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBodyBytes))
	resp, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	respBodyBytes, _ := ioutil.ReadAll(resp.Body)
	newRespHeader := resp.Header
	newRespHeader.Del("Set-Cookie")
	newRespHeader.Del("Date")
	c.record(recordData{
		RequestURLPath:      r.URL.Path,
		RequestMethod:       r.Method,
		RequestHeader:       r.Header,
		RequestBody:         string(reqBodyBytes),
		ResponseHeader:      newRespHeader,
		ResponseBody:        string(respBodyBytes),
		ResponseContentType: resp.Header.Get("Content-Type"),
		ResponseStatus:      resp.Status,
	})
	// Clone resp body
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(respBodyBytes))
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
	apiURL := fmt.Sprintf("/country/annualavg/pr/%s/%s/%s.xml", args.FromCCYY, args.ToCCYY, args.CountryISO)
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

package climate

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"strings"
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
		body := resp.Body
		b, err := ioutil.ReadAll(body)
		if err = xml.NewDecoder(bytes.NewReader(b)).Decode(v); err != nil {
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
	if _, err := c.Do(r, &list); err != nil {
		return List{}, err
	}
	return list, nil
}

func (c *ClientImpl) calculateAveAnnual(list List, fromCCYY, toCCYY int64) (decimal.Decimal, error) {
	domainWebAnnualGcmDatum := list.DomainWebAnnualGcmDatum
	totalAnnualData := decimal.NewFromInt(0)
	totalDatum := int64(len(domainWebAnnualGcmDatum))
	if totalDatum < 1 {
		return decimal.NewFromInt(0), fmt.Errorf("date range %d-%d not supported", fromCCYY, toCCYY)
	}
	totalDatumDec := decimal.NewFromInt(totalDatum)
	for _, v := range domainWebAnnualGcmDatum {
		annualData, err := decimal.NewFromString(v.AnnualData.Double)
		if err != nil {
			continue
		}
		totalAnnualData = totalAnnualData.Add(annualData)
	}
	annualAve := totalAnnualData.Div(totalDatumDec)
	return annualAve, nil
}

func (c *ClientImpl) GetAveAnnualRainfallMany(ctx context.Context, fromCCYY int64, toCCYY int64, countryISOs ...string) (float64, error) {
	var total = float64(0)
	for _, country := range countryISOs {
		result, err := c.GetAveAnnualRainfall(ctx, fromCCYY, toCCYY, country)
		if err != nil {
			return -1, err
		}
		total += result
	}
	return total, nil
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
		return -1, err
	}
	annualAve, err := c.calculateAveAnnual(list, fromCCYY, toCCYY)
	if err != nil {
		return -1, err
	}
	result, _ := annualAve.Float64()
	return result, nil
}

package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/kr/pretty"
	"github.com/tanphamhaiduong/climate_api/climate"
)

func main() {
	validate := validator.New()
	client := climate.NewClient(&http.Client{}, validate, "http://climatedataapi.worldbank.org/climateweb/rest/v1")
	result, err := client.GetAveAnnualRainfall(context.Background(), 1980, 1999, "egy")
	if err != nil {
		log.Fatal(err)
	}
	pretty.Println(result)
}

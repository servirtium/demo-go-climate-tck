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
	client := climate.NewClient(&http.Client{}, validate)
	args := climate.GetAveAnnualRainfallArgs{
		FromCCYY:   "1980",
		ToCCYY:     "1999",
		CountryISO: "GBR",
	}
	result, err := client.GetAveAnnualRainfall(context.Background(), args)
	if err != nil {
		log.Fatal(err)
	}
	pretty.Println(result)
}

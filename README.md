Note - The World Bank took down their climate WebAPI. Darn it. We now depend on a docker version of the same until we work out what to do long term. Docker build and deploy this locally - https://github.com/servirtium/worldbank-climate-recordings - see README

TL;DR:

```
docker build git@github.com:servirtium/worldbank-climate-recordings.git#main -t worldbank-weather-api-for-servirtium-development
docker run -d -p 4567:4567 worldbank-weather-api-for-servirtium-development
```

The build for this demo project needs that docker container running

# climate_api

## How to install vendor modules

Run command `make install`

## How to run tests

Run command `make test`

## How to run tests that directly use the web API

Run command `make test_direct`

## How to run tests that record the web API

Run command `make test_record`

## How to run tests playback previously recorded web API

Run command `make test_playback`

## View Test coverage

Add this command to `~/.profile`  

```  
gocover () {
    t="/tmp/go-cover.$$.tmp"
    go test -coverprofile=$t $@ && go tool cover -html=$t && unlink $t
}
```  
Run `gocover ./...`

## How to run main

Run command `make start`

## How to build

Run command `make build`

## How to remove build file

Run command `make clean`

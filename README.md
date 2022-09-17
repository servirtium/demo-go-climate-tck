Note - The World Bank took down their climate WebAPI. Darn it. We now depend on a docker version of the same until we work out what to do long term. Docker build and deploy this locally - https://github.com/servirtium/worldbank-climate-recordings - see README

TL;DR:

```
docker build git@github.com:servirtium/worldbank-climate-recordings.git#main -t worldbank-weather-api-for-servirtium-development
docker run -d -p 4567:4567 worldbank-weather-api-for-servirtium-development
```

The build for this demo project needs that docker container running.

# Servirtium demo for Go

This repo was build following the step-by-step guide at [https://servirtium.dev/new](https://servirtium.dev/new)

- status: pretty much complete

As well as making a Servirtium library for a language, this step-by-step guide leaves you with a **contrived** example library that serves as a example of how to use Servirtium.

Someone wanting to see an example of how to use Servirtium for a Go-lang project would look at this repo's source. Someone wanting to learn Servirtium by tutorial or extensive reference documentation needs to look elsewhere - sorry!

# Climate API library test harness

A reusable library for Go usage that gives you average rainfall for a country, is what was made to serve as a test harness for this demo. The test harness in turn uses The world bank's REST Web-APIs - `/climateweb/rest/v1/country/annualavg/pr/{fromCCYY}/{toCCYY}/{countryISO}.xml` - for that. See note at top of README.

The demo comes has unit tests and recordings of service interactions for each test.  The recordings are in the [mock](mock) folder.

The library comes with a means to re-record those service interactions, using Servirtium in "record" mode.

Teams evaluating the Servirtium library (but not developing it) would:

* ignore the world bank climate API aspect of this (just for the sake of this demo)
* focus on a HTTP service their application uses (but could easily be outside the dev team in question)
* write their own tests (using their preferred test runner - Mocha, Jasmine, Cucumber.js are fine choices).
* make servirtium optionally do recordings as a mode of operation (commit those recording to Git)
* enjoy their own builds being fast and always green (versus slow and flaky).
* have a non-CI build (daily/weekly) that attempts to re-record and alert that recordings have changed
* remember to keep secrets out of source control (passwords, API tokens and more).

Service tests facilitated by Servirtium (part of the "integration test" class) are one thing, but there should always be a much smaller number of them than **pure unit tests** (no I/O, less than 10ms each). Teams using this library in a larger application would use traditional in-process mocking (say via [GoMock](https://github.com/golang/mock) for the pure unit tests. Reliance on "integration tests" for development on localhost (or far worse a named environment like "dev" or "qa") is a fools game.

Another dev team could use the recordings, as is, to make a new implementation of the library for some reason (say they did not like the license). And they need not even have access to localhost:4567. Some companies happily shipping Servirtium service recordings for specific test scenarios may attach a license agreement that forbids reverse engineering (of the closed-source backend, or the shipped library).

# Building & Tests

## How to install vendor modules

Run command `make install`

## How to run main

Run command `make start`

## How to build

Run command `make build`

## How to remove build file

Run command `make clean`


## How to run tests

Run command `make test`

There are 18 Jest tests in this technology compatibility kit (TCK) project that serves as a demo.

* 6 tests that don't use Servirtium and directly invoke services on WorldBank.com's climate endpoint.
* 6 tests that do the above, but also record the interactions via Servirtium
* 6 tests that don't at all use WorldBank (or need to be online), but instead use the recordings in the above via Servirtium

### Pseudocode for the 6 tests:

```
test_averageRainfallForGreatBritainFrom1980to1999Exists()
    assert climateApi.getAveAnnualRainfall(1980, 1999, "gbr") == 988.8454972331015

test_averageRainfallForFranceFrom1980to1999Exists()
    assert climateApi.getAveAnnualRainfall(1980, 1999, "fra") == 913.7986955122727

test_averageRainfallForEgyptFrom1980to1999Exists()
    assert climateApi.getAveAnnualRainfall(1980, 1999, "egy") == 54.58587712129825

test_averageRainfallForGreatBritainFrom1985to1995DoesNotExist()
    climateApi.getAveAnnualRainfall(1985, 1995, "gbr")
    ... causes "date range not supported" 

test_averageRainfallForMiddleEarthFrom1980to1999DoesNotExist()
    climateApi.getAveAnnualRainfall(1980, 1999, "mde")
    ... causes "bad country code"

test_averageRainfallForGreatBritainAndFranceFrom1980to1999CanBeCalculatedFromTwoRequests()
    assert climateApi.getAveAnnualRainfall(1980, 1999, "gbr", "fra") == 951.3220963726872
```

As mentioned, these six are repeated three times in this test-base: six direct, six record and six playback.

## Running the 6 direct tests only:

Command: `make test_direct`

TODO pic

## Running the 6 tests in record-mode only:

Command: `make test_record`

TODO pic

## Running the 6 tests in playback-mode only:

Command: `make test_playback`

TODO pic

Note the playback mode is quickest. Your day to dey development of you main applications functionality would rely on this mode of operation. 

## Viewing Test coverage

Add this command to `~/.profile`

```  
gocover () {
    t="/tmp/go-cover.$$.tmp"
    go test -coverprofile=$t $@ && go tool cover -html=$t && unlink $t
}
```  
Run `gocover ./...`

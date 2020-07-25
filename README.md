# climate_api
## How to install vendor
Run command `make install`
## How to run test
Run command `make test`
## View Test coverage
Run command  
```
  t="/tmp/go-cover.$$.tmp"
  go test -coverprofile=$t $@ && go tool cover -html=$t && unlink $t
```
## How to run main
Run command `make start`
## How to build
Run command `make build`
## How to remove build file
Run command `make clean`

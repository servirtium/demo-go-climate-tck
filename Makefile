install:
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor
start:
	@go run *.go
build:
	@GOOS=linux GOARCH=amd64 go build -v ./*.go
clean:
	@rm -rf main
test_all:
	@go test -v -race ./...
test_record:
	@go test -v --race -run ^TestClimateTestSuiteRecord$
test_playback:
	@go test -v --race -run ^TestClimateTestSuitePlayback$
test_direct:
	@go test -v --race -run ^TestClimateTestSuiteDirect$

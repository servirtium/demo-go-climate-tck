install:
	@GO111MODULE=on go mod vendor
start:
	@go run *.go
build:
	@GOOS=linux GOARCH=amd64 go build -v ./*.go
clean:
	@rm -rf main
test_all:
	@go test -v -race ./...

TEST_SUITE = "TestClimateTestSuiteRecord"
test:
	@go test -v --race -run ^$(TEST_SUITE)$
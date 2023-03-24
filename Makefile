SKAFFOLD?=skaffold
GO111MODULE?=on
CGO_ENABLED?=0
GOOS?=linux
GOARCH?=amd64
GO_BIN?=app
GO?=go
APP_NAME?=web



build:
	$(MAKE) -C cmd/$(APP_NAME) build
.PHONY=build


mocks: vendor
	$(GO) install github.com/golang/mock/mockgen@v1.6.0
	# generate gomocks
	$(GO) generate ./...
.PHONY=mocks

unit-test: mocks vet
	$(GO) test ./... -p=1 -cover -coverprofile coverage.source.out
	# this will be cached, just needed to the test.json
	$(GO) test ./... -p=1 -cover -coverprofile coverage.source.out -json > test.source.json
	cat coverage.source.out | grep -v "mock_*" | tee coverage.out
	cat test.source.json | grep -v "mock_*" | tee test.json
.PHONY=unit-test

vet:
	$(GO) vet ./...
.PHONY=vet

vendor:
	$(GO) mod vendor
.PHONY=vendor
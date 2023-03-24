.PHONY=build deps

GOPRIVATE?=github.com/wbd-streaming/*
GO111MODULE?=on
CGO_ENABLED?=0
GOOS?=linux
GOARCH?=amd64
GO_BIN?=app
GO?=go
MAIN_DIR?=/var/app

.EXPORT_ALL_VARIABLES:

build:
	$(GO) build -a -installsuffix nocgo -o $(GO_BIN) main.go
BINARY := repokit
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)
CGO_ENABLED ?= 0

.PHONY: build test lint install tidy

build:
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

test:
	go test ./...

lint:
	golangci-lint run

install:
	CGO_ENABLED=$(CGO_ENABLED) go install -ldflags "$(LDFLAGS)" .

tidy:
	go mod tidy

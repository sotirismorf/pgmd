.PHONY: build test test-coverage lint clean install help

BINARY_NAME=pgmd
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

## build: Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/pgmd

## install: Install the binary to $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/pgmd

## test: Run tests
test:
	go test -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linter
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

## fmt: Format code
fmt:
	go fmt ./...
	gofmt -s -w .

## vet: Run go vet
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## tidy: Tidy go modules
tidy:
	go mod tidy

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'

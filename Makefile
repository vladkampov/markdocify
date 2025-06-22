.PHONY: build test clean install lint security-scan

# Build configuration
BINARY_NAME=markdocify
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

# Build the binary
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/markdocify

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run integration tests
test-integration:
	go test -v -tags=integration ./test/integration/...

# Run end-to-end tests
test-e2e:
	go test -v -tags=e2e ./test/e2e/...

# Install locally
install:
	go install $(LDFLAGS) ./cmd/markdocify

# Lint code
lint:
	golangci-lint run

# Security scan
security-scan:
	gosec ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Generate mocks
generate:
	go generate ./...

# Build for all platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/markdocify
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/markdocify
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/markdocify
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/markdocify

# Release preparation
release: clean build-all test
	./scripts/release.sh $(VERSION)
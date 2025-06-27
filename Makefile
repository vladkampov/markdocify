.PHONY: build test clean install lint security-scan fmt fmt-check check-all fmt-fix

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

# Format Go code
fmt:
	@echo "Formatting Go code..."
	@gofmt -w .
	@goimports -w .
	@echo "✓ Code formatted"

# Auto-fix what we can with golangci-lint
fmt-fix:
	@echo "Auto-fixing code issues..."
	@gofmt -w .
	@goimports -w .
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint with auto-fix..."; \
		golangci-lint run --fix; \
	else \
		echo "⚠️  golangci-lint not installed - only running gofmt/goimports"; \
	fi
	@echo "✓ Auto-fixes applied"

# Check if code is formatted
fmt-check:
	@echo "Checking Go code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "❌ The following files need formatting:"; \
		gofmt -l .; \
		echo "Run 'make fmt' to format them"; \
		exit 1; \
	else \
		echo "✓ All files are properly formatted"; \
	fi

# Run all checks locally (same as CI)
check-all: fmt-check
	@echo "Running all local checks..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running linter..."; \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "   brew install golangci-lint"; \
		echo "   OR: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi
	@echo "Running tests..."
	@go test ./...
	@echo "✅ All checks passed!"

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
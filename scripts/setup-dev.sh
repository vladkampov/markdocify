#!/bin/bash
# Development environment setup script

set -e

echo "🔧 Setting up development environment..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

echo "✅ Go is installed: $(go version)"

# Install development tools
echo "📦 Installing development tools..."

# Install golangci-lint if not present
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    if command -v brew &> /dev/null; then
        brew install golangci-lint
    else
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi
    echo "✅ golangci-lint installed"
else
    echo "✅ golangci-lint already installed"
fi

# Install goimports if not present
if ! command -v goimports &> /dev/null; then
    echo "Installing goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
    echo "✅ goimports installed"
else
    echo "✅ goimports already installed"
fi

# Install gosec for security scanning
if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    echo "✅ gosec installed"
else
    echo "✅ gosec already installed"
fi

# Download dependencies
echo "📥 Downloading Go modules..."
go mod download
go mod verify

# Run initial checks
echo "🧪 Running initial checks..."
make fmt-check
make check-all

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "Useful commands:"
echo "  make check-all    - Run all checks (formatting, linting, tests)"
echo "  make fmt-fix      - Auto-fix formatting and some lint issues"
echo "  make fmt          - Format code with gofmt and goimports"
echo "  make lint         - Run golangci-lint"
echo "  make test         - Run tests"
echo ""
echo "The pre-commit hook is already installed and will check:"
echo "  ✓ Go code formatting (always)"
echo "  ✓ Linting (if golangci-lint is installed)"
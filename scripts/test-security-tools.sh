#!/bin/bash

# Test security tools locally before pushing to CI
# This helps verify the tools work correctly

set -e

echo "🔐 Testing Security Tools Locally"
echo "================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

test_tool() {
    local tool_name=$1
    local test_command=$2
    
    echo -e "\n🔍 Testing ${YELLOW}$tool_name${NC}..."
    
    if eval "$test_command"; then
        echo -e "✅ ${GREEN}$tool_name${NC}: Working correctly"
        return 0
    else
        echo -e "❌ ${RED}$tool_name${NC}: Failed"
        return 1
    fi
}

# Test Go installation
test_tool "Go" "go version"

# Test go vet (built-in security analysis)
test_tool "go vet" "go vet ./..."

# Test staticcheck
echo -e "\n📦 Installing staticcheck..."
go install honnef.co/go/tools/cmd/staticcheck@latest

test_tool "staticcheck" "~/go/bin/staticcheck ./..."

# Test nancy
echo -e "\n📦 Installing nancy..."
go install github.com/sonatypecommunity/nancy@latest

test_tool "nancy" "go list -json -deps ./... | nancy sleuth"

# Test govulncheck
echo -e "\n📦 Installing govulncheck..."
go install golang.org/x/vuln/cmd/govulncheck@latest

test_tool "govulncheck" "~/go/bin/govulncheck ./..."

# Test basic build
test_tool "go build" "go build -o test-markdocify ./cmd/markdocify"

# Cleanup
echo -e "\n🧹 Cleaning up..."
rm -f test-markdocify

echo -e "\n🎉 ${GREEN}All security tools tested successfully!${NC}"
echo ""
echo "Your CI/CD pipeline should now work correctly."
echo "Tools verified:"
echo "  ✅ go vet (built-in static analysis)"
echo "  ✅ staticcheck (advanced static analysis)"
echo "  ✅ nancy (vulnerability scanner)"
echo "  ✅ govulncheck (Go vulnerability scanner)"
echo "  ✅ go build (compilation test)"

exit 0
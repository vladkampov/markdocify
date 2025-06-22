#!/bin/bash

# markdocify CI/CD Setup Verification Script
# This script verifies that your CI/CD pipeline is properly configured

set -e

echo "🔍 markdocify CI/CD Setup Verification"
echo "======================================"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Success/failure tracking
SUCCESS=0
TOTAL=0

check_file() {
    local file=$1
    local description=$2
    TOTAL=$((TOTAL + 1))
    
    if [[ -f "$file" ]]; then
        echo -e "✅ ${GREEN}$description${NC}: $file"
        SUCCESS=$((SUCCESS + 1))
    else
        echo -e "❌ ${RED}$description${NC}: $file (missing)"
    fi
}

check_directory() {
    local dir=$1
    local description=$2
    TOTAL=$((TOTAL + 1))
    
    if [[ -d "$dir" ]]; then
        echo -e "✅ ${GREEN}$description${NC}: $dir"
        SUCCESS=$((SUCCESS + 1))
    else
        echo -e "❌ ${RED}$description${NC}: $dir (missing)"
    fi
}

echo ""
echo "📁 Checking CI/CD Files..."

# GitHub workflows
check_directory ".github" "GitHub directory"
check_directory ".github/workflows" "Workflows directory"
check_file ".github/workflows/ci.yml" "CI workflow"
check_file ".github/workflows/release.yml" "Release workflow"
check_file ".github/workflows/security.yml" "Security workflow"
check_file ".github/dependabot.yml" "Dependabot configuration"

# Build configuration
check_file ".goreleaser.yml" "GoReleaser configuration"
check_file "Dockerfile" "Docker configuration"
check_file "Makefile" "Build configuration"

# Documentation
check_file "CI_CD_SETUP.md" "CI/CD setup guide"
check_file "README.md" "Project README"
check_file "CONTRIBUTING.md" "Contributing guide"
check_file "LICENSE" "License file"

echo ""
echo "🔧 Checking Go Project Structure..."

# Go project files
check_file "go.mod" "Go module file"
check_file "go.sum" "Go dependencies"
check_directory "cmd/markdocify" "Main command directory"
check_directory "internal" "Internal packages"

echo ""
echo "📊 Checking Code Quality..."

# Test if go mod is valid
TOTAL=$((TOTAL + 1))
if go mod verify > /dev/null 2>&1; then
    echo -e "✅ ${GREEN}Go module verification${NC}: valid"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "❌ ${RED}Go module verification${NC}: invalid"
fi

# Test if project builds
TOTAL=$((TOTAL + 1))
if go build ./cmd/markdocify > /dev/null 2>&1; then
    echo -e "✅ ${GREEN}Go build test${NC}: successful"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "❌ ${RED}Go build test${NC}: failed"
fi

# Test if tests pass
TOTAL=$((TOTAL + 1))
if go test ./... > /dev/null 2>&1; then
    echo -e "✅ ${GREEN}Go test suite${NC}: passing"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "❌ ${RED}Go test suite${NC}: failing"
fi

echo ""
echo "🔐 Security Checks..."

# Check for security tools
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

TOTAL=$((TOTAL + 1))
if command_exists gosec; then
    echo -e "✅ ${GREEN}gosec security scanner${NC}: installed"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "⚠️  ${YELLOW}gosec security scanner${NC}: not installed (will be installed in CI)"
fi

TOTAL=$((TOTAL + 1))
if command_exists nancy; then
    echo -e "✅ ${GREEN}nancy vulnerability scanner${NC}: installed"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "⚠️  ${YELLOW}nancy vulnerability scanner${NC}: not installed (will be installed in CI)"
fi

echo ""
echo "🏷️  Git Configuration..."

# Check if we're in a git repository
TOTAL=$((TOTAL + 1))
if git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "✅ ${GREEN}Git repository${NC}: initialized"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "❌ ${RED}Git repository${NC}: not initialized"
fi

# Check for git origin
TOTAL=$((TOTAL + 1))
if git remote get-url origin > /dev/null 2>&1; then
    ORIGIN=$(git remote get-url origin)
    echo -e "✅ ${GREEN}Git origin${NC}: $ORIGIN"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "❌ ${RED}Git origin${NC}: not configured"
fi

echo ""
echo "📋 Setup Status Summary"
echo "======================"

PERCENTAGE=$((SUCCESS * 100 / TOTAL))

if [[ $SUCCESS -eq $TOTAL ]]; then
    echo -e "🎉 ${GREEN}Perfect! All checks passed ($SUCCESS/$TOTAL)${NC}"
    echo ""
    echo "Your CI/CD pipeline is ready! 🚀"
    echo ""
    echo "Next steps:"
    echo "1. Push your code to GitHub"
    echo "2. Configure repository secrets (see CI_CD_SETUP.md)"
    echo "3. Create your first release with: git tag v1.0.0 && git push origin v1.0.0"
elif [[ $PERCENTAGE -ge 80 ]]; then
    echo -e "✅ ${GREEN}Great! Most checks passed ($SUCCESS/$TOTAL - $PERCENTAGE%)${NC}"
    echo ""
    echo "Your setup is mostly ready. Review failed checks above."
elif [[ $PERCENTAGE -ge 60 ]]; then
    echo -e "⚠️  ${YELLOW}Good progress ($SUCCESS/$TOTAL - $PERCENTAGE%)${NC}"
    echo ""
    echo "Several items need attention. See failed checks above."
else
    echo -e "❌ ${RED}Setup incomplete ($SUCCESS/$TOTAL - $PERCENTAGE%)${NC}"
    echo ""
    echo "Please address the failed checks above before proceeding."
fi

echo ""
echo "📖 For detailed setup instructions, see: CI_CD_SETUP.md"
echo "🔗 GitHub Actions: https://github.com/vladkampov/markdocify/actions"

exit 0
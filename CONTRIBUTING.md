# Contributing to markdocify

Welcome to markdocify! We're excited that you're interested in contributing. This guide will help you get started with contributing to the project.

## 🎯 Vision

markdocify aims to be the **go-to tool for converting documentation websites into high-quality Markdown**. We focus on:

- **Comprehensive coverage**: Capturing complete documentation hierarchies
- **Intelligent extraction**: Smart content detection across different frameworks
- **Clean output**: LLM-ready, well-formatted Markdown
- **Performance**: Fast, efficient scraping with respectful practices
- **Usability**: Zero-config operation for most documentation sites

## 🚀 Quick Start for Contributors

### Prerequisites

- **Go 1.21+** (required)
- **Make** (recommended for build automation)
- **Git** (for version control)
- **golangci-lint** (for code quality checks)

### Development Setup

1. **Fork and Clone**
```bash
git clone https://github.com/your-username/markdocify.git
cd markdocify
```

2. **Install Dependencies**
```bash
go mod tidy
```

3. **Build and Test**
```bash
# Build the project
make build

# Run tests
make test

# Run linting
make lint

# Test locally
./bin/markdocify --help
```

4. **Verify Everything Works**
```bash
# Quick test on a simple site
./bin/markdocify https://example.com -d 1 -o test.md
cat test.md  # Should show example.com content
```

## 🛠️ Development Workflow

### Making Changes

1. **Create a Feature Branch**
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-description
```

2. **Make Your Changes**
   - Write clear, well-documented code
   - Follow Go conventions and best practices
   - Add tests for new functionality
   - Update documentation as needed

3. **Test Thoroughly**
```bash
# Run all tests
make test

# Test your changes manually
./bin/markdocify https://docs.example.com -d 2

# Run linting
make lint

# Check for security issues
make security-scan
```

4. **Commit with Clear Messages**
```bash
git add .
git commit -m "feat: add support for custom content selectors"
# or
git commit -m "fix: handle edge case in URL parsing"
```

### Commit Message Format

We follow conventional commits:

- `feat:` New features
- `fix:` Bug fixes  
- `docs:` Documentation changes
- `test:` Adding or fixing tests
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `chore:` Maintenance tasks

Examples:
```
feat: add chromedp support for javascript-rendered sites
fix: resolve infinite loop in link following
docs: update installation instructions
test: add integration tests for config loading
```

## 🎯 Areas Where We Need Help

### 🔥 High Priority

1. **JavaScript Rendering Support**
   - Integrate ChromeDP for SPA documentation sites
   - Add configuration options for browser automation
   - Handle dynamic content loading

2. **Enhanced Content Detection**
   - Add selectors for more documentation frameworks
   - Improve content quality scoring
   - Better handling of code examples and API references

3. **Performance Improvements**
   - Optimize memory usage for large sites
   - Implement smarter caching strategies
   - Parallel processing improvements

### 🌟 Medium Priority

4. **Output Formats**
   - JSON output for structured data
   - HTML export with styling
   - Custom template support

5. **Advanced Configuration**
   - Per-domain configuration profiles
   - Plugin system for custom processors
   - Environment variable support

6. **Testing & Quality**
   - Integration tests with real documentation sites
   - Performance benchmarks
   - Edge case coverage

### 💡 Ideas Welcome

7. **New Features**
   - API for programmatic usage
   - Web interface for non-technical users
   - Integration with documentation platforms
   - Support for authentication-protected docs

## 🧪 Testing Guidelines

### Test Types

1. **Unit Tests**
```bash
# Run specific package tests
go test ./internal/config -v
go test ./internal/scraper -v
```

2. **Integration Tests**
```bash
# Test with real sites (when available)
make test-integration
```

3. **Manual Testing**
```bash
# Test with different documentation sites
./bin/markdocify https://docs.python.org/3/ -d 2
./bin/markdocify https://nextjs.org/docs -d 3
```

### Writing Tests

- **Add tests for all new functions**
- **Use table-driven tests for multiple scenarios**
- **Mock external dependencies**
- **Test edge cases and error conditions**

Example test structure:
```go
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name        string
        config      Config
        expectError bool
        errorMsg    string
    }{
        {
            name: "valid config",
            config: Config{
                Name: "Test",
                BaseURL: "https://example.com",
                // ...
            },
            expectError: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 📝 Code Style Guidelines

### Go Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Run `golangci-lint` before submitting
- Write clear, self-documenting code

### Documentation

- **Add GoDoc comments** for all public functions and types
- **Update README.md** for user-facing changes
- **Include examples** in documentation
- **Comment complex logic** inline

### Error Handling

- **Use meaningful error messages**
- **Wrap errors with context**: `fmt.Errorf("failed to parse config: %w", err)`
- **Handle all error cases**
- **Don't ignore errors**

Example:
```go
func ProcessURL(url string) error {
    if url == "" {
        return fmt.Errorf("URL cannot be empty")
    }
    
    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("failed to fetch URL %s: %w", url, err)
    }
    defer resp.Body.Close()
    
    // Process response...
    return nil
}
```

## 🐛 Bug Reports

### Before Reporting

1. **Check existing issues** to avoid duplicates
2. **Test with latest version**
3. **Try with minimal reproduction case**

### Bug Report Template

```markdown
**Description**
A clear description of the bug.

**To Reproduce**
Steps to reproduce the behavior:
1. Run command: `markdocify https://example.com/docs`
2. Observe error/unexpected behavior

**Expected Behavior**
What you expected to happen.

**Actual Behavior**
What actually happened.

**Environment**
- OS: [e.g., Ubuntu 20.04, macOS 12.1, Windows 10]
- markdocify version: [e.g., v1.2.3]
- Go version: [e.g., 1.21.0]

**Additional Context**
- Config file (if used)
- Log output
- Screenshots (if applicable)
```

## 🌟 Feature Requests

### Before Requesting

1. **Check if similar feature exists**
2. **Consider if it fits the project scope**
3. **Think about implementation approach**

### Feature Request Template

```markdown
**Feature Description**
Clear description of the proposed feature.

**Use Case**
Why is this feature needed? What problem does it solve?

**Proposed Solution**
How should this feature work?

**Alternatives Considered**
Other approaches you've considered.

**Additional Context**
Any other relevant information.
```

## 🔄 Pull Request Process

### Before Submitting

1. **Ensure all tests pass**: `make test`
2. **Run linting**: `make lint`
3. **Test manually** with real documentation sites
4. **Update documentation** if needed
5. **Add/update tests** for your changes

### PR Template

```markdown
**Description**
Brief description of changes.

**Type of Change**
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

**Testing**
- [ ] Tests pass locally
- [ ] Added tests for new functionality
- [ ] Manually tested with documentation sites

**Checklist**
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or clearly documented)
```

### Review Process

1. **Automated checks** must pass (tests, linting)
2. **Code review** by maintainers
3. **Testing** with real documentation sites
4. **Approval** and merge

## 🏗️ Project Architecture

### Directory Structure

```
markdocify/
├── cmd/markdocify/              # CLI application entry point
├── internal/
│   ├── config/                 # Configuration parsing and validation
│   ├── scraper/                # Web scraping engine (Colly-based)
│   ├── converter/              # HTML to Markdown conversion
│   ├── aggregator/             # Document aggregation and TOC
│   ├── types/                  # Shared type definitions
│   ├── monitor/                # Metrics and monitoring (future)
│   └── legal/                  # Robots.txt and compliance (future)
├── configs/examples/           # Example configuration files
├── test/
│   ├── unit/                   # Unit tests
│   ├── integration/            # Integration tests
│   └── e2e/                    # End-to-end tests
└── docs/                       # Additional documentation
```

### Key Components

1. **CLI (`cmd/markdocify/`)**: Command-line interface and argument parsing
2. **Scraper (`internal/scraper/`)**: Web scraping logic using Colly
3. **Converter (`internal/converter/`)**: HTML to Markdown conversion
4. **Aggregator (`internal/aggregator/`)**: Combines pages into single document
5. **Config (`internal/config/`)**: Configuration handling and validation

### Data Flow

```
URL Input → Config → Scraper → Converter → Aggregator → Markdown Output
    ↓           ↓        ↓         ↓           ↓
 Validation  Patterns  HTML   Markdown    TOC & Meta
```

## 🤝 Community Guidelines

### Code of Conduct

- **Be respectful** and inclusive
- **Focus on constructive feedback**
- **Help others learn and grow**
- **Assume good intentions**

### Getting Help

- **Read the documentation** first (README, this guide)
- **Search existing issues** before asking
- **Provide context** when asking questions
- **Be patient** - maintainers are volunteers

### Recognition

Contributors are recognized in:
- **Contributors list** in README
- **Release notes** for significant contributions
- **Special mentions** for major features

## 📞 Contact

- **Issues**: [GitHub Issues](https://github.com/vladkampov/markdocify/issues)
- **Discussions**: [GitHub Discussions](https://github.com/vladkampov/markdocify/discussions)
- **Email**: For security issues only

---

**Thank you for contributing to markdocify!** 🎉

Every contribution, no matter how small, helps make documentation more accessible to everyone.
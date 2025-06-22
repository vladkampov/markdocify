# 📚 markdocify

> **Comprehensively scrape documentation sites into beautiful, LLM-ready Markdown**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/vladkampov/markdocify.svg)](https://github.com/vladkampov/markdocify/releases)
[![CI](https://github.com/vladkampov/markdocify/workflows/CI/badge.svg)](https://github.com/vladkampov/markdocify/actions/workflows/ci.yml)
[![Security](https://github.com/vladkampov/markdocify/workflows/Security%20Scan/badge.svg)](https://github.com/vladkampov/markdocify/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/vladkampov/markdocify/branch/main/graph/badge.svg)](https://codecov.io/gh/vladkampov/markdocify)
[![Go Report Card](https://goreportcard.com/badge/github.com/vladkampov/markdocify)](https://goreportcard.com/report/github.com/vladkampov/markdocify)

markdocify is a powerful CLI tool that **comprehensively scrapes documentation websites** and converts them into well-formatted, single Markdown files. Perfect for creating LLM training data, offline documentation, or comprehensive knowledge bases.

## ✨ Features

- 🎯 **Comprehensive Coverage**: Scrapes deep hierarchical documentation (8 levels by default)
- 🧠 **Intelligent Content Detection**: Auto-detects documentation patterns across popular frameworks
- 🚫 **Smart Filtering**: Automatically excludes navigation, ads, and non-documentation content
- ⚡ **High Performance**: Concurrent scraping with configurable workers and delays
- 📊 **Progress Reporting**: Real-time progress updates for long scrapes
- 🔧 **Zero Configuration**: Works out-of-the-box for most documentation sites
- 🎨 **Clean Output**: Generates well-formatted Markdown with table of contents
- 🛡️ **Respectful Scraping**: Built-in rate limiting and robots.txt compliance

## 🚀 Quick Start

### Installation

**🍺 Homebrew (macOS/Linux)** - Recommended
```bash
# Add our tap and install
brew tap vladkampov/tap
brew install markdocify

# Or install directly
brew install vladkampov/tap/markdocify
```

**📦 Package Managers**
```bash
# Snap (Ubuntu/Linux)
sudo snap install markdocify

# Scoop (Windows)
scoop bucket add vladkampov https://github.com/vladkampov/scoop-bucket
scoop install markdocify

# AUR (Arch Linux)
yay -S markdocify-bin
```

**⬇️ Direct Download** 
```bash
# Download latest release for your platform
curl -L https://github.com/vladkampov/markdocify/releases/latest/download/markdocify-linux-amd64 -o markdocify
chmod +x markdocify

# Or for macOS
curl -L https://github.com/vladkampov/markdocify/releases/latest/download/markdocify-darwin-amd64 -o markdocify
chmod +x markdocify
```

**🐳 Docker**
```bash
# Run directly with Docker
docker run --rm -v $(pwd):/workspace ghcr.io/vladkampov/markdocify:latest https://example.com/docs

# Or use as base image
FROM ghcr.io/vladkampov/markdocify:latest
```

**🔧 Build from Source**
```bash
git clone https://github.com/vladkampov/markdocify.git
cd markdocify
make build
```

**Go Install**
```bash
go install github.com/vladkampov/markdocify/cmd/markdocify@latest
```

### Basic Usage

```bash
# Comprehensive scrape (recommended) - captures full documentation
markdocify https://vercel.com/docs

# Quick scrape - lighter, faster
markdocify https://docs.example.com -d 3

# Custom output file
markdocify https://react.dev/docs -o react-complete-docs.md

# Adjust performance settings
markdocify https://site.com/docs -d 5 --concurrency 4
```

## 💡 Use Cases

### 📖 LLM Training Data
Create comprehensive, clean Markdown datasets from documentation sites:
```bash
markdocify https://nextjs.org/docs -o nextjs-training-data.md
markdocify https://docs.python.org -o python-docs.md  
markdocify https://kubernetes.io/docs -o k8s-complete.md
```

### 📚 Offline Documentation
Generate complete offline documentation archives:
```bash
markdocify https://docs.aws.amazon.com/ec2 -o aws-ec2-offline.md
markdocify https://tailwindcss.com/docs -o tailwind-offline.md
```

### 🔍 Knowledge Bases
Create searchable, comprehensive knowledge bases:
```bash
markdocify https://docs.github.com -o github-docs-complete.md
markdocify https://api.stripe.com/docs -o stripe-api-complete.md
```

## 🎯 Supported Sites

markdocify works great with most documentation sites, including:

- **Frameworks**: React, Vue, Angular, Next.js, Nuxt, SvelteKit, Astro
- **Platforms**: Vercel, Netlify, AWS, Google Cloud, Azure
- **Languages**: Python, Go, Rust, JavaScript, TypeScript docs
- **Tools**: Docker, Kubernetes, Terraform, GitHub, GitLab
- **Databases**: PostgreSQL, MongoDB, Redis documentation
- **And many more!**

## ⚙️ Configuration

### Command Line Options

```bash
markdocify [URL] [flags]

Flags:
  -c, --config string      Configuration file path
  -o, --output string      Output file path  
  -d, --depth int          Maximum crawl depth (default 8)
      --concurrency int    Number of concurrent workers (default 3)
  -h, --help              Help for markdocify
  -v, --version           Version information
```

### Advanced Configuration

For complex sites, use YAML configuration files:

```yaml
# custom-config.yml
name: "Custom Documentation"
base_url: "https://example.com"
output_file: "custom-docs.md"

start_urls:
  - "https://example.com/docs"
  - "https://example.com/api"

follow_patterns:
  - "^https://example\\.com/docs/.*"
  - "^https://example\\.com/api/.*"

processing:
  max_depth: 10
  concurrency: 5
  delay: 0.5
  preserve_code_blocks: true
  generate_toc: true

selectors:
  title: "h1, .page-title"
  content: "main, .documentation"
  exclude:
    - "nav"
    - ".sidebar"
    - "footer"
```

Use with: `markdocify -c custom-config.yml`

## 📊 Performance & Output

### Typical Results

| Site | Pages Scraped | Output Size | Time |
|------|---------------|-------------|------|
| Vercel Docs | 100+ pages | 2-5MB | 3-5 min |
| Next.js Docs | 80+ pages | 1-3MB | 2-4 min |
| React Docs | 50+ pages | 800KB-2MB | 1-3 min |

### Output Quality

markdocify generates:
- 📑 **Table of Contents** with deep linking
- 🏷️ **Metadata** including source URLs and timestamps  
- 🎨 **Clean formatting** with preserved code blocks
- 🔗 **Resolved links** and proper heading hierarchy
- 🧹 **Filtered content** with navigation/ads removed

## 🛠️ Development

### Prerequisites
- Go 1.21+
- Make

### Building
```bash
# Clone repository
git clone https://github.com/vladkampov/markdocify.git
cd markdocify

# Download dependencies
go mod tidy

# Build
make build

# Run tests
make test

# Cross-platform build
make build-all
```

### Project Structure
```
markdocify/
├── cmd/markdocify/          # CLI application
├── internal/
│   ├── config/             # Configuration handling
│   ├── scraper/            # Web scraping engine
│   ├── converter/          # HTML to Markdown conversion
│   ├── aggregator/         # Document aggregation & TOC
│   └── types/              # Shared types
├── configs/examples/       # Example configurations
└── README.md
```

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Quick Contribution Guide

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Make** your changes with tests
4. **Test** thoroughly: `make test && make lint`
5. **Commit** with clear messages
6. **Submit** a pull request

### Areas We Need Help

- 🌐 **JavaScript rendering** support (ChromeDP integration)
- 🔍 **More content selectors** for different documentation frameworks
- 🎨 **Output formats** (JSON, HTML, etc.)
- 🚀 **Performance optimizations**
- 📚 **Documentation** improvements
- 🧪 **Test coverage** expansion

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Colly](https://github.com/gocolly/colly) for web scraping
- Powered by [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) for conversion
- CLI built with [Cobra](https://github.com/spf13/cobra)
- Inspired by the need for high-quality LLM training data

## 📞 Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/vladkampov/markdocify/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/vladkampov/markdocify/discussions)
- 📖 **Documentation**: [Project Wiki](https://github.com/vladkampov/markdocify/wiki)

---

<p align="center">
  <strong>Made with ❤️ for the developer community</strong><br>
  Star ⭐ this repo if you find it useful!
</p>
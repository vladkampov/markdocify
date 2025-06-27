package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/vladkampov/markdocify/internal/config"
	"github.com/vladkampov/markdocify/internal/scraper"
)

var version = "dev" // Will be overridden by ldflags during build

var rootCmd = &cobra.Command{
	Use:   "markdocify [URL]",
	Short: "Comprehensively scrape documentation sites into Markdown",
	Long: `markdocify is a CLI tool that comprehensively scrapes documentation websites
and converts them into a single, well-formatted Markdown file.

It aggressively follows documentation links to capture complete documentation sites,
with intelligent content detection and smart filtering of non-documentation content.

Usage:
  markdocify -c config.yml                    # Use configuration file
  markdocify https://example.com/docs         # Comprehensive scrape (depth 8)
  markdocify https://example.com/docs -o out.md  # Custom output file
  markdocify https://example.com/docs -d 5    # Custom depth (lighter scrape)`,
	Version: version,
	RunE:    runScraper,
}

var configFile string
var outputFile string
var maxDepth int
var concurrency int

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Configuration file path")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output file path")
	rootCmd.PersistentFlags().IntVarP(&maxDepth, "depth", "d", 8, "Maximum crawl depth (for URL mode)")
	rootCmd.PersistentFlags().IntVar(&concurrency, "concurrency", 3, "Number of concurrent workers (for URL mode)")
}

func runScraper(cmd *cobra.Command, args []string) error {
	var cfg *config.Config
	var err error

	// Check if URL is provided as argument (quick mode)
	if len(args) > 0 && configFile == "" {
		url := args[0]
		cfg, err = createQuickConfig(url)
		if err != nil {
			return fmt.Errorf("failed to create quick config: %w", err)
		}
	} else if configFile != "" {
		// Use configuration file
		cfg, err = config.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	} else {
		return fmt.Errorf("either provide a URL as argument or use -c flag with configuration file")
	}

	if outputFile != "" {
		cfg.OutputFile = outputFile
	}

	scraperInstance, err := scraper.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create scraper: %w", err)
	}

	if err := scraperInstance.Run(); err != nil {
		return fmt.Errorf("scraping failed: %w", err)
	}

	fmt.Printf("Successfully scraped documentation to %s\n", cfg.OutputFile)
	return nil
}

func createQuickConfig(inputURL string) (*config.Config, error) {
	// Validate URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		inputURL = "https://" + inputURL
		parsedURL, err = url.Parse(inputURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL after adding https: %w", err)
		}
	}

	// Generate a reasonable output filename
	if outputFile == "" {
		hostname := parsedURL.Hostname()
		hostname = strings.ReplaceAll(hostname, ".", "-")
		outputFile = fmt.Sprintf("%s-docs.md", hostname)
	}

	// Create comprehensive domain patterns for following links
	domain := parsedURL.Hostname()
	basePattern := strings.ReplaceAll(domain, ".", "\\.")

	// Much more aggressive following patterns for comprehensive documentation coverage
	followPatterns := []string{
		fmt.Sprintf("^https?://%s/.*", basePattern), // Main domain pattern
	}

	// Add specific documentation path patterns based on the starting URL
	if strings.Contains(inputURL, "/docs") {
		followPatterns = append(followPatterns,
			fmt.Sprintf("^https?://%s/docs/.*", basePattern),
			fmt.Sprintf("^https?://%s/documentation/.*", basePattern),
			fmt.Sprintf("^https?://%s/guide/.*", basePattern),
			fmt.Sprintf("^https?://%s/guides/.*", basePattern),
			fmt.Sprintf("^https?://%s/tutorial/.*", basePattern),
			fmt.Sprintf("^https?://%s/tutorials/.*", basePattern),
			fmt.Sprintf("^https?://%s/reference/.*", basePattern),
			fmt.Sprintf("^https?://%s/api/.*", basePattern),
			fmt.Sprintf("^https?://%s/cli/.*", basePattern),
			fmt.Sprintf("^https?://%s/learn/.*", basePattern),
		)
	}

	// Create quick configuration
	cfg := &config.Config{
		Name:       fmt.Sprintf("%s Documentation", titleCase(domain)),
		BaseURL:    fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host),
		OutputFile: outputFile,
		StartURLs:  []string{inputURL},

		FollowPatterns: followPatterns,
		IgnorePatterns: []string{
			// Media files
			".*\\.(jpg|jpeg|png|gif|svg|css|js|ico|woff|woff2|ttf|eot|pdf|zip|tar|gz)$",
			// Non-documentation pages
			".*/edit.*$",
			".*/settings.*$",
			".*/login.*$",
			".*/logout.*$",
			".*/signin.*$",
			".*/signup.*$",
			".*/register.*$",
			".*/contact.*$",
			".*/support.*$",
			".*/pricing.*$",
			".*/about.*$",
			".*/careers.*$",
			".*/jobs.*$",
			".*/legal.*$",
			".*/terms.*$",
			".*/privacy.*$",
			// Social and external
			".*github\\.com.*$",
			".*twitter\\.com.*$",
			".*linkedin\\.com.*$",
			".*facebook\\.com.*$",
			".*youtube\\.com.*$",
			".*discord\\.(gg|com).*$",
			".*slack\\.com.*$",
			// Admin and user-specific
			".*/admin.*$",
			".*/dashboard.*$",
			".*/account.*$",
			".*/profile.*$",
			".*/user/.*$",
			// Interactive features that aren't documentation
			".*/playground.*$",
			".*/editor.*$",
			".*/sandbox.*$",
		},

		Selectors: config.SelectorConfig{
			Title: "h1, title, .page-title, .doc-title, [data-testid='page-title']",
			Content: strings.Join([]string{
				// Primary content containers
				"main", "article", ".content", ".documentation", ".docs", "#content", ".main-content",
				// Documentation-specific containers
				".doc-content", ".docs-content", ".documentation-content", ".guide-content",
				".tutorial-content", ".reference-content", ".api-content", ".markdown-body",
				// Framework-specific patterns
				".docusaurus_skipToContent_node", ".nextra-content", ".vuepress-content",
				".gitbook-content", ".notion-page-content", ".sphinx-content",
				// Generic content patterns
				"[role='main']", "[data-content]", ".page-content", ".post-content",
				".entry-content", ".single-content", "#main-content", "#primary-content",
				// Fallback to body if nothing else matches
				"body",
			}, ", "),
			Exclude: []string{
				// Navigation elements
				"nav", "header", "footer", ".navigation", ".sidebar", ".toc", ".table-of-contents",
				".menu", ".nav", ".navbar", ".topbar", ".breadcrumb", ".breadcrumbs",
				// Interactive elements
				".edit-link", ".edit-page", ".edit-this-page", ".feedback", ".prev-next",
				".pagination", ".page-nav", ".site-nav", ".social", ".share",
				// Code/technical elements to exclude from main content
				"script", "style", "noscript", ".highlight", ".code-toolbar",
				// Ads and tracking
				".advertisement", ".ads", ".ad", ".promo", ".banner", ".cookie",
				// Comments and social
				".comments", ".disqus", ".utterances", ".giscus", ".social-share",
				// Search and forms
				".search", ".search-box", ".search-form", ".newsletter", ".subscribe",
				// Specific known patterns from popular doc sites
				".cmdklaunch_wrapper__KrfZL", ".mobile-menu_root__PX9iM", ".toggle_mobileMenuToggle__W5y02",
				".header_header__TSZx7", "[data-testid='header']", ".DocSearch", ".algolia-docsearch",
				// Version/language switchers
				".version-switcher", ".lang-switcher", ".theme-switcher", ".locale-switcher",
			},
		},

		Processing: config.ProcessingConfig{
			MaxDepth:           maxDepth,
			Concurrency:        concurrency,
			Delay:              0.8, // Slightly faster for comprehensive scraping
			PreserveCodeBlocks: true,
			GenerateTOC:        true,
			SanitizeHTML:       true,
		},

		Engines: []config.EngineConfig{
			{
				Type:      "colly",
				UserAgent: "markdocify/1.0 (+https://github.com/vladkampov/markdocify)",
				Timeout:   30,
			},
		},

		Output: config.OutputConfig{
			HeadingOffset:      0,
			IncludeMetadata:    true,
			SyntaxHighlighting: true,
			PreserveImages:     false,
			InlineStyles:       false,
		},

		Security: config.SecurityConfig{
			RespectRobots:    true,
			CheckTerms:       false,
			MaxFileSize:      "10MB",
			AllowedDomains:   []string{domain},
			MaxFileSizeBytes: 10 * 1024 * 1024,
		},

		Monitoring: config.MonitoringConfig{
			EnableMetrics:   false,
			LogLevel:        "info",
			ProgressUpdates: true,
			MetricsPort:     9090,
		},
	}

	// Set defaults and validate
	if err := cfg.SetDefaults(); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	fmt.Printf("Quick scrape mode: %s -> %s\n", inputURL, outputFile)
	fmt.Printf("Max depth: %d, Concurrency: %d\n", maxDepth, concurrency)

	return cfg, nil
}

func titleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = string(unicode.ToUpper(rune(word[0]))) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

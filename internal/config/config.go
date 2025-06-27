package config

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name       string `yaml:"name" validate:"required"`
	BaseURL    string `yaml:"base_url" validate:"required,url"`
	OutputFile string `yaml:"output_file" validate:"required"`

	StartURLs      []string `yaml:"start_urls" validate:"required,min=1"`
	FollowPatterns []string `yaml:"follow_patterns"`
	IgnorePatterns []string `yaml:"ignore_patterns"`

	Selectors SelectorConfig `yaml:"selectors"`

	Processing ProcessingConfig `yaml:"processing"`
	Engines    []EngineConfig   `yaml:"engines"`
	Output     OutputConfig     `yaml:"output"`
	Security   SecurityConfig   `yaml:"security"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
}

type SelectorConfig struct {
	Title      string   `yaml:"title"`
	Content    string   `yaml:"content" validate:"required"`
	Navigation string   `yaml:"navigation"`
	Exclude    []string `yaml:"exclude"`
}

type ProcessingConfig struct {
	MaxDepth           int     `yaml:"max_depth"`
	Concurrency        int     `yaml:"concurrency"`
	Delay              float64 `yaml:"delay"`
	PreserveCodeBlocks bool    `yaml:"preserve_code_blocks"`
	GenerateTOC        bool    `yaml:"generate_toc"`
	SanitizeHTML       bool    `yaml:"sanitize_html"`
}

type EngineConfig struct {
	Type         string `yaml:"type" validate:"required,oneof=colly chromedp"`
	UserAgent    string `yaml:"user_agent"`
	Timeout      int    `yaml:"timeout"`
	WaitSelector string `yaml:"wait_selector"`
}

type OutputConfig struct {
	HeadingOffset      int  `yaml:"heading_offset"`
	IncludeMetadata    bool `yaml:"include_metadata"`
	SyntaxHighlighting bool `yaml:"syntax_highlighting"`
	PreserveImages     bool `yaml:"preserve_images"`
	InlineStyles       bool `yaml:"inline_styles"`
}

type SecurityConfig struct {
	RespectRobots    bool          `yaml:"respect_robots"`
	CheckTerms       bool          `yaml:"check_terms"`
	MaxFileSize      string        `yaml:"max_file_size"`
	AllowedDomains   []string      `yaml:"allowed_domains"`
	RequestTimeout   time.Duration `yaml:"request_timeout"`
	ScrapingTimeout  time.Duration `yaml:"scraping_timeout"`
	MaxFileSizeBytes int64
}

type MonitoringConfig struct {
	EnableMetrics   bool   `yaml:"enable_metrics"`
	LogLevel        string `yaml:"log_level"`
	ProgressUpdates bool   `yaml:"progress_updates"`
	MetricsPort     int    `yaml:"metrics_port"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.SetDefaults(); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func (c *Config) SetDefaults() error {
	if c.Processing.MaxDepth == 0 {
		c.Processing.MaxDepth = 5
	}
	if c.Processing.Concurrency == 0 {
		c.Processing.Concurrency = 3
	}
	if c.Processing.Delay == 0 {
		c.Processing.Delay = 1.0
	}
	if c.Selectors.Title == "" {
		c.Selectors.Title = "h1"
	}
	if c.Selectors.Content == "" {
		c.Selectors.Content = "main, article, .content"
	}
	if c.Monitoring.LogLevel == "" {
		c.Monitoring.LogLevel = "info"
	}
	if c.Monitoring.MetricsPort == 0 {
		c.Monitoring.MetricsPort = 9090
	}
	if c.Security.MaxFileSize == "" {
		c.Security.MaxFileSize = "10MB"
	}

	maxSize, err := parseSize(c.Security.MaxFileSize)
	if err != nil {
		return fmt.Errorf("invalid max_file_size: %w", err)
	}
	c.Security.MaxFileSizeBytes = maxSize
	c.Security.RequestTimeout = 30 * time.Second

	// Set default scraping timeout - generous for large documentation sites
	if c.Security.ScrapingTimeout == 0 {
		c.Security.ScrapingTimeout = 10 * time.Minute
	}

	if len(c.Engines) == 0 {
		c.Engines = []EngineConfig{
			{
				Type:      "colly",
				UserAgent: "docs-scraper/1.0",
				Timeout:   30,
			},
		}
	}

	return nil
}

func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if err := validateURL(c.BaseURL, "base_url"); err != nil {
		return err
	}

	if c.OutputFile == "" {
		return fmt.Errorf("output_file is required")
	}

	if len(c.StartURLs) == 0 {
		return fmt.Errorf("start_urls is required and must contain at least one URL")
	}

	// Validate all start URLs
	for i, startURL := range c.StartURLs {
		if err := validateURL(startURL, fmt.Sprintf("start_urls[%d]", i)); err != nil {
			return err
		}
	}

	// Validate regex patterns
	for i, pattern := range c.FollowPatterns {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid follow_pattern[%d] '%s': %w", i, pattern, err)
		}
	}

	for i, pattern := range c.IgnorePatterns {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid ignore_pattern[%d] '%s': %w", i, pattern, err)
		}
	}

	// Validate processing configuration
	if c.Processing.MaxDepth <= 0 {
		return fmt.Errorf("max_depth must be greater than 0, got %d", c.Processing.MaxDepth)
	}
	if c.Processing.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than 0, got %d", c.Processing.Concurrency)
	}
	if c.Processing.Delay < 0 {
		return fmt.Errorf("delay must be non-negative, got %f", c.Processing.Delay)
	}

	// Validate allowed domains if specified
	for i, domain := range c.Security.AllowedDomains {
		if domain == "" {
			return fmt.Errorf("allowed_domains[%d] cannot be empty", i)
		}
		// Basic domain format validation
		if len(domain) < 3 || !regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(domain) {
			return fmt.Errorf("invalid domain format in allowed_domains[%d]: '%s'", i, domain)
		}
	}

	return nil
}

// validateURL validates that a URL string is well-formed and contains required components
func validateURL(urlStr, fieldName string) error {
	if urlStr == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid %s '%s': %w", fieldName, urlStr, err)
	}

	if parsed.Scheme == "" {
		return fmt.Errorf("%s must include scheme (http/https): '%s'", fieldName, urlStr)
	}

	if parsed.Host == "" {
		return fmt.Errorf("%s must include host: '%s'", fieldName, urlStr)
	}

	// Ensure scheme is http or https
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s must use http or https scheme: '%s'", fieldName, urlStr)
	}

	return nil
}

func parseSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("size string is empty")
	}

	var multiplier int64 = 1
	var numStr string

	switch {
	case len(sizeStr) >= 2 && sizeStr[len(sizeStr)-2:] == "KB":
		multiplier = 1024
		numStr = sizeStr[:len(sizeStr)-2]
	case len(sizeStr) >= 2 && sizeStr[len(sizeStr)-2:] == "MB":
		multiplier = 1024 * 1024
		numStr = sizeStr[:len(sizeStr)-2]
	case len(sizeStr) >= 2 && sizeStr[len(sizeStr)-2:] == "GB":
		multiplier = 1024 * 1024 * 1024
		numStr = sizeStr[:len(sizeStr)-2]
	default:
		numStr = sizeStr
	}

	var num int64
	if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
		return 0, fmt.Errorf("invalid size format: %w", err)
	}

	return num * multiplier, nil
}

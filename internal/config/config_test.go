package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
name: "Test Documentation"
base_url: "https://example.com"
output_file: "test-output.md"

start_urls:
  - "https://example.com/docs"

selectors:
  title: "h1"
  content: "main"

processing:
  max_depth: 3
  concurrency: 2
  delay: 1.0
  preserve_code_blocks: true
  generate_toc: true
  sanitize_html: true

security:
  respect_robots: true
  max_file_size: "5MB"
  allowed_domains:
    - "example.com"
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Test loading the config
	config, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	// Verify the loaded configuration
	assert.Equal(t, "Test Documentation", config.Name)
	assert.Equal(t, "https://example.com", config.BaseURL)
	assert.Equal(t, "test-output.md", config.OutputFile)
	assert.Equal(t, []string{"https://example.com/docs"}, config.StartURLs)
	assert.Equal(t, "h1", config.Selectors.Title)
	assert.Equal(t, "main", config.Selectors.Content)
	assert.Equal(t, 3, config.Processing.MaxDepth)
	assert.Equal(t, 2, config.Processing.Concurrency)
	assert.Equal(t, 1.0, config.Processing.Delay)
	assert.True(t, config.Processing.PreserveCodeBlocks)
	assert.True(t, config.Processing.GenerateTOC)
	assert.True(t, config.Processing.SanitizeHTML)
	assert.True(t, config.Security.RespectRobots)
	assert.Equal(t, []string{"example.com"}, config.Security.AllowedDomains)
	assert.Equal(t, int64(5*1024*1024), config.Security.MaxFileSizeBytes)
}

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
				Name:       "Test",
				BaseURL:    "https://example.com",
				OutputFile: "output.md",
				StartURLs:  []string{"https://example.com/docs"},
				Processing: ProcessingConfig{
					MaxDepth:    5,
					Concurrency: 3,
				},
			},
			expectError: false,
		},
		{
			name: "missing name",
			config: Config{
				BaseURL:    "https://example.com",
				OutputFile: "output.md",
				StartURLs:  []string{"https://example.com/docs"},
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "invalid base_url",
			config: Config{
				Name:       "Test",
				BaseURL:    "not-a-url",
				OutputFile: "output.md",
				StartURLs:  []string{"https://example.com/docs"},
				Processing: ProcessingConfig{MaxDepth: 1, Concurrency: 1},
			},
			expectError: true,
			errorMsg:    "base_url must include scheme",
		},
		{
			name: "invalid start URL",
			config: Config{
				Name:       "Test",
				BaseURL:    "https://example.com",
				OutputFile: "output.md",
				StartURLs:  []string{"not-a-url"},
				Processing: ProcessingConfig{MaxDepth: 1, Concurrency: 1},
			},
			expectError: true,
			errorMsg:    "start_urls[0] must include scheme",
		},
		{
			name: "empty start_urls",
			config: Config{
				Name:       "Test",
				BaseURL:    "https://example.com",
				OutputFile: "output.md",
				StartURLs:  []string{},
			},
			expectError: true,
			errorMsg:    "start_urls is required",
		},
		{
			name: "invalid follow pattern",
			config: Config{
				Name:           "Test",
				BaseURL:        "https://example.com",
				OutputFile:     "output.md",
				StartURLs:      []string{"https://example.com/docs"},
				FollowPatterns: []string{"[invalid regex"},
				Processing: ProcessingConfig{
					MaxDepth:    5,
					Concurrency: 3,
				},
			},
			expectError: true,
			errorMsg:    "invalid follow_pattern[0]",
		},
		{
			name: "invalid allowed domain",
			config: Config{
				Name:       "Test",
				BaseURL:    "https://example.com",
				OutputFile: "output.md",
				StartURLs:  []string{"https://example.com/docs"},
				Processing: ProcessingConfig{MaxDepth: 1, Concurrency: 1},
				Security: SecurityConfig{
					AllowedDomains: []string{"invalid-domain"},
				},
			},
			expectError: true,
			errorMsg:    "invalid domain format",
		},
		{
			name: "zero max depth",
			config: Config{
				Name:       "Test",
				BaseURL:    "https://example.com",
				OutputFile: "output.md",
				StartURLs:  []string{"https://example.com/docs"},
				Processing: ProcessingConfig{MaxDepth: 0, Concurrency: 1},
			},
			expectError: true,
			errorMsg:    "max_depth must be greater than 0, got 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"1024", 1024, false},
		{"1KB", 1024, false},
		{"1MB", 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"10MB", 10 * 1024 * 1024, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseSize(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
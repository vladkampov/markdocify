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

func TestLoadConfig_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (string, func())
		expectError string
	}{
		{
			name: "nonexistent file",
			setupFunc: func() (string, func()) {
				return "/nonexistent/path/config.yml", func() {}
			},
			expectError: "failed to read config file",
		},
		{
			name: "invalid YAML",
			setupFunc: func() (string, func()) {
				tmpFile, _ := os.CreateTemp("", "invalid-*.yml")
				_, _ = tmpFile.WriteString("invalid: yaml: content: [")
				tmpFile.Close()
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectError: "failed to parse config file",
		},
		{
			name: "validation failure",
			setupFunc: func() (string, func()) {
				tmpFile, _ := os.CreateTemp("", "invalid-config-*.yml")
				_, _ = tmpFile.WriteString(`
name: ""
base_url: "invalid-url"
output_file: "test.md"
start_urls: []
`)
				tmpFile.Close()
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectError: "invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, cleanup := tt.setupFunc()
			defer cleanup()

			_, err := LoadConfig(configPath)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestSetDefaults_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedError  string
		validateResult func(*testing.T, *Config)
	}{
		{
			name: "invalid max file size",
			config: Config{
				Security: SecurityConfig{
					MaxFileSize: "invalid-size",
				},
			},
			expectedError: "invalid max_file_size",
		},
		{
			name: "all defaults applied",
			config: Config{
				Security: SecurityConfig{
					MaxFileSize: "5MB",
				},
			},
			validateResult: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 5, cfg.Processing.MaxDepth)
				assert.Equal(t, 3, cfg.Processing.Concurrency)
				assert.Equal(t, 1.0, cfg.Processing.Delay)
				assert.Equal(t, "h1", cfg.Selectors.Title)
				assert.Equal(t, "main, article, .content", cfg.Selectors.Content)
				assert.Equal(t, "info", cfg.Monitoring.LogLevel)
				assert.Equal(t, 9090, cfg.Monitoring.MetricsPort)
				assert.Equal(t, int64(5*1024*1024), cfg.Security.MaxFileSizeBytes)
				assert.Equal(t, "colly", cfg.Engines[0].Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.SetDefaults()

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, &tt.config)
				}
			}
		})
	}
}

func TestValidateURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		fieldName   string
		expectError string
	}{
		{
			name:        "empty URL",
			url:         "",
			fieldName:   "test_url",
			expectError: "test_url cannot be empty",
		},
		{
			name:        "malformed URL",
			url:         "ht tp://invalid url",
			fieldName:   "test_url",
			expectError: "invalid test_url",
		},
		{
			name:        "missing scheme",
			url:         "example.com/path",
			fieldName:   "test_url",
			expectError: "test_url must include scheme",
		},
		{
			name:        "missing host",
			url:         "https://",
			fieldName:   "test_url",
			expectError: "test_url must include host",
		},
		{
			name:        "invalid scheme",
			url:         "ftp://example.com",
			fieldName:   "test_url",
			expectError: "test_url must use http or https scheme",
		},
		{
			name:        "valid https URL",
			url:         "https://example.com/path",
			fieldName:   "test_url",
			expectError: "",
		},
		{
			name:        "valid http URL",
			url:         "http://example.com",
			fieldName:   "test_url",
			expectError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url, tt.fieldName)

			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_ComprehensiveCases(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError string
	}{
		{
			name: "negative delay",
			config: Config{
				Name:       "Test",
				BaseURL:    "https://example.com",
				OutputFile: "test.md",
				StartURLs:  []string{"https://example.com"},
				Processing: ProcessingConfig{
					MaxDepth:    1,
					Concurrency: 1,
					Delay:       -1.0,
				},
			},
			expectError: "delay must be non-negative",
		},
		{
			name: "negative concurrency",
			config: Config{
				Name:       "Test",
				BaseURL:    "https://example.com",
				OutputFile: "test.md",
				StartURLs:  []string{"https://example.com"},
				Processing: ProcessingConfig{
					MaxDepth:    1,
					Concurrency: 0,
					Delay:       1.0,
				},
			},
			expectError: "concurrency must be greater than 0",
		},
		{
			name: "empty allowed domain",
			config: Config{
				Name:       "Test",
				BaseURL:    "https://example.com",
				OutputFile: "test.md",
				StartURLs:  []string{"https://example.com"},
				Processing: ProcessingConfig{
					MaxDepth:    1,
					Concurrency: 1,
				},
				Security: SecurityConfig{
					AllowedDomains: []string{""},
				},
			},
			expectError: "allowed_domains[0] cannot be empty",
		},
		{
			name: "invalid domain format",
			config: Config{
				Name:       "Test",
				BaseURL:    "https://example.com",
				OutputFile: "test.md",
				StartURLs:  []string{"https://example.com"},
				Processing: ProcessingConfig{
					MaxDepth:    1,
					Concurrency: 1,
				},
				Security: SecurityConfig{
					AllowedDomains: []string{"invalid-domain-format"},
				},
			},
			expectError: "invalid domain format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

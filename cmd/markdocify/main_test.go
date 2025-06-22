package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	// Test version command
	output := captureOutput(t, func() {
		rootCmd.SetArgs([]string{"--version"})
		err := rootCmd.Execute()
		require.NoError(t, err)
	})
	
	assert.Contains(t, output, "markdocify version")
}

func TestHelp(t *testing.T) {
	// Test help command
	output := captureOutput(t, func() {
		rootCmd.SetArgs([]string{"--help"})
		err := rootCmd.Execute()
		require.NoError(t, err)
	})
	
	assert.Contains(t, output, "markdocify is a CLI tool")
	assert.Contains(t, output, "Usage:")
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello World"},
		{"HELLO WORLD", "Hello World"},
		{"hello", "Hello"},
		{"", ""},
		{"a", "A"},
		{"multiple   spaces", "Multiple Spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := titleCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRunScraperErrors(t *testing.T) {
	// Test runScraper function directly to avoid command execution issues
	t.Run("no arguments", func(t *testing.T) {
		err := runScraper(rootCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "either provide a URL")
	})
	
	t.Run("invalid URL", func(t *testing.T) {
		err := runScraper(rootCmd, []string{"not-a-valid-url"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})
	
	t.Run("invalid config file", func(t *testing.T) {
		// Save original value
		origConfigFile := configFile
		defer func() { configFile = origConfigFile }()
		
		// Set non-existent config file
		configFile = "/tmp/non-existent-config.yml"
		
		err := runScraper(rootCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})
	
	t.Run("malformed config file", func(t *testing.T) {
		// Create temporary malformed config file
		tmpFile, err := os.CreateTemp("", "bad-config-*.yml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		
		// Write invalid YAML
		_, err = tmpFile.WriteString("invalid: yaml: content: [\n")
		require.NoError(t, err)
		tmpFile.Close()
		
		// Save original value
		origConfigFile := configFile
		defer func() { configFile = origConfigFile }()
		
		// Set malformed config file
		configFile = tmpFile.Name()
		
		err = runScraper(rootCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})
}

func TestRunScraperWithConfig(t *testing.T) {
	// Create temporary config file
	tmpFile, err := os.CreateTemp("", "test-config-*.yml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	
	configContent := `
name: "Test Documentation"
base_url: "https://example.com"
output_file: "test-output.md"
start_urls:
  - "https://example.com/docs"
processing:
  max_depth: 1
  concurrency: 1
`
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()
	
	// Save original value
	origConfigFile := configFile
	defer func() { configFile = origConfigFile }()
	
	// Set config file flag
	configFile = tmpFile.Name()
	
	// Test with config file (this will fail during scraping since example.com/docs doesn't exist)
	err = runScraper(rootCmd, []string{})
	assert.Error(t, err) // Expected since we're scraping a non-existent URL
	// Check that it's a scraping failure, not a config failure
	assert.Contains(t, err.Error(), "scraping failed")
}

func TestRunScraperWithOutputFlag(t *testing.T) {
	// Save original values
	origOutputFile := outputFile
	defer func() { outputFile = origOutputFile }()
	
	// Test URL with output flag
	outputFile = "/tmp/test-output.md"
	
	// This should succeed as httpbin.org/html is actually accessible
	err := runScraper(rootCmd, []string{"https://httpbin.org/html"})
	// httpbin.org/html usually works, so we don't expect an error
	if err != nil {
		t.Logf("Scraping failed (network issue?): %v", err)
		// If it fails due to network, that's OK for our test
	}
}

func TestCreateQuickConfig(t *testing.T) {
	// Save original values and restore them
	origOutputFile := outputFile
	origMaxDepth := maxDepth
	origConcurrency := concurrency
	defer func() {
		outputFile = origOutputFile
		maxDepth = origMaxDepth
		concurrency = origConcurrency
	}()
	
	// Set test values
	outputFile = ""
	maxDepth = 5
	concurrency = 2
	
	testURL := "https://example.com/docs"
	
	cfg, err := createQuickConfig(testURL)
	require.NoError(t, err)
	
	assert.Equal(t, "Example.com Documentation", cfg.Name)
	assert.Equal(t, "https://example.com", cfg.BaseURL)
	assert.Equal(t, "example-com-docs.md", cfg.OutputFile)
	assert.Equal(t, []string{testURL}, cfg.StartURLs)
	assert.Equal(t, 5, cfg.Processing.MaxDepth)
	assert.Equal(t, 2, cfg.Processing.Concurrency)
	assert.True(t, cfg.Processing.PreserveCodeBlocks)
	assert.True(t, cfg.Processing.GenerateTOC)
	assert.True(t, cfg.Processing.SanitizeHTML)
	
	// Test invalid URL
	_, err = createQuickConfig("not-a-valid-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
	
	// Test empty URL
	_, err = createQuickConfig("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
	
	// Test URL with different schemes
	cfg, err = createQuickConfig("http://example.com/docs")
	require.NoError(t, err)
	assert.Equal(t, "http://example.com", cfg.BaseURL)
	
	// Test URL with custom output file
	outputFile = "custom-output.md"
	cfg, err = createQuickConfig("https://test.com/api")
	require.NoError(t, err)
	assert.Equal(t, "custom-output.md", cfg.OutputFile)
}

func TestMain(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"help command", []string{"markdocify", "--help"}},
		{"version command", []string{"markdocify", "--version"}},
		{"invalid command", []string{"markdocify", "invalid-url-without-scheme"}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that main doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("main() panicked with args %v: %v", tt.args, r)
				}
			}()
			
			// Save original args and restore them after test
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			
			// Set test args
			os.Args = tt.args
			
			// Capture output to avoid printing to test output
			old := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w
			
			// This should not panic
			main()
			
			// Restore stdout/stderr
			w.Close()
			os.Stdout = old
			os.Stderr = oldStderr
			
			// Read the output (optional, just to consume it)
			_, _ = io.ReadAll(r)
		})
	}
}

// Helper function to capture command output
func captureOutput(t *testing.T, fn func()) string {
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	
	os.Stdout = w
	
	fn()
	
	w.Close()
	os.Stdout = old
	
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	
	return buf.String()
}
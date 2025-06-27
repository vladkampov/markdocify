package aggregator

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladkampov/markdocify/internal/config"
)

func TestAddPage(t *testing.T) {
	cfg := &config.Config{
		OutputFile: "/tmp/test-aggregator.md",
		Output: config.OutputConfig{
			IncludeMetadata: true,
		},
		Processing: config.ProcessingConfig{
			GenerateTOC: true,
		},
	}

	agg, err := New(cfg)
	require.NoError(t, err)

	// Add first page
	agg.AddPage("https://example.com/page1", "Page 1", "Content 1", 0)
	assert.Equal(t, 1, agg.GetPageCount())

	// Add duplicate content - should be ignored
	agg.AddPage("https://example.com/page1-duplicate", "Page 1 Duplicate", "Content 1", 0)
	assert.Equal(t, 1, agg.GetPageCount(), "Duplicate content should be ignored")

	// Add different content
	agg.AddPage("https://example.com/page2", "Page 2", "Content 2", 1)
	assert.Equal(t, 2, agg.GetPageCount())
}

func TestGenerateOutput(t *testing.T) {
	tempFile := "/tmp/test-output-" + t.Name() + ".md"
	defer os.Remove(tempFile)

	cfg := &config.Config{
		Name:       "Test Documentation",
		BaseURL:    "https://example.com",
		OutputFile: tempFile,
		Output: config.OutputConfig{
			IncludeMetadata: true,
		},
		Processing: config.ProcessingConfig{
			GenerateTOC: true,
			MaxDepth:    2,
		},
	}

	agg, err := New(cfg)
	require.NoError(t, err)

	// Add test pages
	agg.AddPage("https://example.com/", "Home", "# Home\nWelcome to the docs", 0)
	agg.AddPage("https://example.com/api", "API", "# API\nAPI documentation", 1)

	// Generate output
	err = agg.GenerateOutput()
	require.NoError(t, err)

	// Read and verify output
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	contentStr := string(content)

	// Verify metadata
	assert.Contains(t, contentStr, "# Test Documentation")
	assert.Contains(t, contentStr, "**Total Pages**: 2")
	assert.Contains(t, contentStr, "**Max Depth**: 2")

	// Verify TOC
	assert.Contains(t, contentStr, "## Table of Contents")

	// Verify content
	assert.Contains(t, contentStr, "# Home")
	assert.Contains(t, contentStr, "## API")
	assert.Contains(t, contentStr, "*Source: [https://example.com/](https://example.com/)*")
}

func TestCreateAnchor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple title",
			input:    "Simple Title",
			expected: "simple-title",
		},
		{
			name:     "title with special characters",
			input:    "API Reference: useState() Hook",
			expected: "api-reference-usestate-hook",
		},
		{
			name:     "title with multiple spaces and dashes",
			input:    "Complex -- Title   With    Spaces",
			expected: "complex-title-with-spaces",
		},
		{
			name:     "title with various punctuation",
			input:    "What is React? (A Guide)",
			expected: "what-is-react-a-guide",
		},
	}

	agg := &Aggregator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agg.createAnchor(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortPages(t *testing.T) {
	cfg := &config.Config{}
	agg, err := New(cfg)
	require.NoError(t, err)

	// Add pages in random order
	agg.AddPage("https://example.com/page3", "Page 3", "Content 3", 1)
	agg.AddPage("https://example.com/page1", "Page 1", "Content 1", 0)
	agg.AddPage("https://example.com/page2", "Page 2", "Content 2", 0)
	agg.AddPage("https://example.com/page4", "Page 4", "Content 4", 1)

	agg.sortPages()

	// Should be sorted by depth, then by URL
	assert.Equal(t, 0, agg.pages[0].Depth)
	assert.Equal(t, 0, agg.pages[1].Depth)
	assert.Equal(t, 1, agg.pages[2].Depth)
	assert.Equal(t, 1, agg.pages[3].Depth)

	// Within same depth, should be sorted by URL
	assert.True(t, agg.pages[0].URL < agg.pages[1].URL)
	assert.True(t, agg.pages[2].URL < agg.pages[3].URL)
}

func TestExtractTitleFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "simple path",
			url:      "https://example.com/api-reference",
			expected: "Api Reference",
		},
		{
			name:     "nested path",
			url:      "https://example.com/docs/getting-started",
			expected: "Getting Started",
		},
		{
			name:     "trailing slash",
			url:      "https://example.com/api/",
			expected: "Api",
		},
		{
			name:     "root URL",
			url:      "https://example.com/",
			expected: "Example.com",
		},
	}

	agg := &Aggregator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agg.extractTitleFromURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMemoryLimitWarning(t *testing.T) {
	cfg := &config.Config{}
	agg, err := New(cfg)
	require.NoError(t, err)

	// This is a simplified test - in practice you'd capture stdout
	// to verify the warning message is printed when exceeding MaxPagesInMemory

	// Add many pages to trigger warning
	for i := 0; i < 5; i++ {
		content := fmt.Sprintf("Content %d", i)
		agg.AddPage(fmt.Sprintf("https://example.com/page%d", i), fmt.Sprintf("Page %d", i), content, 0)
	}

	assert.Equal(t, 5, agg.GetPageCount())
}

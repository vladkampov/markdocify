package converter

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladkampov/markdocify/internal/config"
	"github.com/vladkampov/markdocify/internal/types"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Processing: config.ProcessingConfig{
			SanitizeHTML: true,
		},
		Output: config.OutputConfig{
			IncludeMetadata: true,
		},
	}

	converter, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, converter.sanitizer)
	assert.NotNil(t, converter.mdConverter)
	assert.Equal(t, cfg, converter.config)
}

func TestCreateSanitizer(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		inputHTML      string
		expectContains []string
		expectRemoved  []string
	}{
		{
			name: "basic sanitization",
			config: &config.Config{
				Output: config.OutputConfig{
					PreserveImages: false,
					InlineStyles:   false,
				},
			},
			inputHTML:      `<script>alert('xss')</script><p>Safe content</p>`,
			expectContains: []string{"Safe content"},
			expectRemoved:  []string{"script", "alert"},
		},
		{
			name: "preserve images",
			config: &config.Config{
				Output: config.OutputConfig{
					PreserveImages: true,
					InlineStyles:   false,
				},
			},
			inputHTML:      `<img src="test.jpg" alt="test"><p>Content</p>`,
			expectContains: []string{"img", "src=\"test.jpg\"", "alt=\"test\""},
		},
		{
			name: "code blocks preserved",
			config: &config.Config{
				Output: config.OutputConfig{
					PreserveImages: false,
					InlineStyles:   false,
				},
			},
			inputHTML:      `<pre><code class="language-go">func main() {}</code></pre>`,
			expectContains: []string{"<pre>", "<code", "func main()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := &Converter{config: tt.config}
			sanitizer := converter.createSanitizer()

			result := sanitizer.Sanitize(tt.inputHTML)

			for _, expected := range tt.expectContains {
				assert.Contains(t, result, expected)
			}

			for _, removed := range tt.expectRemoved {
				assert.NotContains(t, result, removed)
			}
		})
	}
}

func TestCreateMarkdownConverter(t *testing.T) {
	converter := &Converter{
		config: &config.Config{},
	}

	mdConverter := converter.createMarkdownConverter()
	assert.NotNil(t, mdConverter)

	// Test that it can convert basic HTML
	result, err := mdConverter.ConvertString("<h1>Title</h1><p>Paragraph</p>")
	require.NoError(t, err)
	assert.Contains(t, result, "# Title")
	assert.Contains(t, result, "Paragraph")
}

func TestConvertToMarkdown(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		pageContent  *types.PageContent
		expectError  bool
		expectResult []string
	}{
		{
			name: "basic conversion",
			config: &config.Config{
				Processing: config.ProcessingConfig{
					SanitizeHTML: true,
				},
				Output: config.OutputConfig{
					IncludeMetadata: false,
				},
			},
			pageContent: &types.PageContent{
				URL:     "https://example.com",
				Title:   "Test Page",
				Content: "<h1>Title</h1><p>Content here</p>",
				Depth:   1,
			},
			expectError:  false,
			expectResult: []string{"# Title", "Content here"},
		},
		{
			name: "with metadata",
			config: &config.Config{
				Processing: config.ProcessingConfig{
					SanitizeHTML: true,
				},
				Output: config.OutputConfig{
					IncludeMetadata: true,
				},
			},
			pageContent: &types.PageContent{
				URL:     "https://example.com/page",
				Title:   "Test Page",
				Content: "<p>Content</p>",
				Depth:   2,
			},
			expectError: false,
			expectResult: []string{
				"<!-- Source: https://example.com/page -->",
				"<!-- Title: Test Page -->",
				"<!-- Depth: 2 -->",
				"Content",
			},
		},
		{
			name: "empty content",
			config: &config.Config{
				Processing: config.ProcessingConfig{
					SanitizeHTML: true,
				},
			},
			pageContent: &types.PageContent{
				URL:     "https://example.com",
				Title:   "Empty",
				Content: "",
				Depth:   1,
			},
			expectError: true,
		},
		{
			name: "sanitization disabled",
			config: &config.Config{
				Processing: config.ProcessingConfig{
					SanitizeHTML: false,
				},
				Output: config.OutputConfig{
					IncludeMetadata: false,
				},
			},
			pageContent: &types.PageContent{
				URL:     "https://example.com",
				Title:   "Test",
				Content: "<script>alert('test')</script><p>Content</p>",
				Depth:   1,
			},
			expectError:  false,
			expectResult: []string{"Content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter, err := New(tt.config)
			require.NoError(t, err)

			result, err := converter.ConvertToMarkdown(tt.pageContent)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, expected := range tt.expectResult {
					assert.Contains(t, result, expected)
				}
			}
		})
	}
}

func TestPostProcessMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove trailing whitespace",
			input:    "Line 1   \nLine 2\t\n",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "consolidate empty lines",
			input:    "Line 1\n\n\n\nLine 2",
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "remove trailing empty lines",
			input:    "Content\n\n\n",
			expected: "Content",
		},
		{
			name:     "preserve single empty line",
			input:    "Line 1\n\nLine 2",
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "handle only whitespace",
			input:    "   \t\n   \n",
			expected: "",
		},
	}

	converter := &Converter{config: &config.Config{}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.postProcessMarkdown(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateMetadata(t *testing.T) {
	converter := &Converter{config: &config.Config{}}

	page := &types.PageContent{
		URL:       "https://example.com/test",
		Title:     "Test Page",
		Content:   "Some content",
		Depth:     3,
		Timestamp: time.Now(),
	}

	metadata := converter.generateMetadata(page)

	expectedLines := []string{
		"<!-- Source: https://example.com/test -->",
		"<!-- Title: Test Page -->",
		"<!-- Depth: 3 -->",
	}

	for _, expected := range expectedLines {
		assert.Contains(t, metadata, expected)
	}

	// Check that all lines are present
	lines := strings.Split(metadata, "\n")
	assert.Len(t, lines, 3)
}

func TestConvertToMarkdown_ErrorCases(t *testing.T) {
	converter, err := New(&config.Config{
		Processing: config.ProcessingConfig{
			SanitizeHTML: true,
		},
	})
	require.NoError(t, err)

	// Test invalid HTML that might cause conversion errors
	page := &types.PageContent{
		URL:     "https://example.com",
		Title:   "Test",
		Content: "<p>Valid content</p>",
		Depth:   1,
	}

	result, err := converter.ConvertToMarkdown(page)
	assert.NoError(t, err)
	assert.Contains(t, result, "Valid content")
}

func TestSanitizer_ComplexHTML(t *testing.T) {
	config := &config.Config{
		Output: config.OutputConfig{
			PreserveImages: true,
			InlineStyles:   false,
		},
	}

	converter := &Converter{config: config}
	sanitizer := converter.createSanitizer()

	complexHTML := `
		<div class="container">
			<h1>Title</h1>
			<p>Paragraph with <strong>bold</strong> and <em>italic</em></p>
			<pre><code class="language-go">func test() {}</code></pre>
			<ul>
				<li>Item 1</li>
				<li>Item 2</li>
			</ul>
			<table>
				<thead>
					<tr><th>Header</th></tr>
				</thead>
				<tbody>
					<tr><td>Cell</td></tr>
				</tbody>
			</table>
			<img src="test.jpg" alt="Test image">
			<a href="https://example.com" title="Link">Link text</a>
			<script>alert('malicious')</script>
		</div>
	`

	result := sanitizer.Sanitize(complexHTML)

	// Should preserve safe elements
	assert.Contains(t, result, "<h1>Title</h1>")
	assert.Contains(t, result, "<strong>bold</strong>")
	assert.Contains(t, result, "<pre><code")
	assert.Contains(t, result, "<ul>")
	assert.Contains(t, result, "<table>")
	assert.Contains(t, result, `<img src="test.jpg" alt="Test image">`)
	assert.Contains(t, result, `<a href="https://example.com"`)

	// Should remove dangerous elements
	assert.NotContains(t, result, "<script>")
	assert.NotContains(t, result, "alert('malicious')")
}

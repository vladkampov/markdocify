package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladkampov/markdocify/internal/config"
)

func TestIsAllowedDomain(t *testing.T) {
	tests := []struct {
		name           string
		allowedDomains []string
		url            string
		expected       bool
	}{
		{
			name:           "exact domain match",
			allowedDomains: []string{"example.com"},
			url:            "https://example.com/page",
			expected:       true,
		},
		{
			name:           "subdomain match",
			allowedDomains: []string{"example.com"},
			url:            "https://docs.example.com/page",
			expected:       true,
		},
		{
			name:           "prevent subdomain injection",
			allowedDomains: []string{"example.com"},
			url:            "https://maliciousexample.com/page",
			expected:       false,
		},
		{
			name:           "no allowed domains means allow all",
			allowedDomains: []string{},
			url:            "https://anywhere.com/page",
			expected:       true,
		},
		{
			name:           "invalid URL",
			allowedDomains: []string{"example.com"},
			url:            "not-a-url",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Security: config.SecurityConfig{
					AllowedDomains: tt.allowedDomains,
				},
			}
			scraper := &Scraper{config: cfg}

			result := scraper.isAllowedDomain(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPrivacyOrLegalURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "privacy policy URL",
			url:      "https://example.com/privacy",
			expected: true,
		},
		{
			name:     "terms URL",
			url:      "https://example.com/terms",
			expected: true,
		},
		{
			name:     "about page",
			url:      "https://example.com/about",
			expected: true,
		},
		{
			name:     "documentation URL",
			url:      "https://example.com/docs/api",
			expected: false,
		},
		{
			name:     "blog post URL should not be skipped",
			url:      "https://example.com/blog/post",
			expected: false,
		},
		{
			name:     "careers page",
			url:      "https://example.com/careers",
			expected: true,
		},
	}

	scraper := &Scraper{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scraper.isPrivacyOrLegalURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove React branding",
			input:    "useState – React",
			expected: "useState",
		},
		{
			name:     "remove Stripe branding",
			input:    "Balance – Stripe",
			expected: "Balance",
		},
		{
			name:     "remove repeated words",
			input:    "API API Documentation",
			expected: "API Documentation",
		},
		{
			name:     "clean complex title",
			input:    "useEffect – React This feature is available in the latest Experimental version of React",
			expected: "useEffect",
		},
		{
			name:     "simple title unchanged",
			input:    "Simple Title",
			expected: "Simple Title",
		},
	}

	scraper := &Scraper{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scraper.cleanTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScraperIntegration(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`
<!DOCTYPE html>
<html>
<head><title>Test Documentation</title></head>
<body>
	<main>
		<h1>Test Documentation</h1>
		<p>This is test content.</p>
		<a href="/page1">Page 1</a>
	</main>
</body>
</html>
			`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/page1":
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`
<!DOCTYPE html>
<html>
<head><title>Page 1</title></head>
<body>
	<main>
		<h1>Page 1</h1>
		<p>This is page 1 content.</p>
	</main>
</body>
</html>
			`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create test configuration
	cfg := &config.Config{
		Name:       "Test Documentation",
		BaseURL:    server.URL,
		OutputFile: "/tmp/test-output.md",
		StartURLs:  []string{server.URL},
		Processing: config.ProcessingConfig{
			MaxDepth:           2,
			Concurrency:        1,
			Delay:              0.1,
			PreserveCodeBlocks: true,
			GenerateTOC:        true,
		},
		Security: config.SecurityConfig{
			RequestTimeout:  10 * time.Second,
			ScrapingTimeout: 30 * time.Second,
		},
		Monitoring: config.MonitoringConfig{
			LogLevel: "error", // Reduce noise in tests
		},
	}

	// Initialize scraper
	scraper, err := New(cfg)
	require.NoError(t, err)

	// Run scraper
	err = scraper.Run()
	assert.NoError(t, err)

	// Verify results
	pageCount := scraper.aggregator.GetPageCount()
	assert.Greater(t, pageCount, 0, "Should have scraped at least one page")
}

func TestVisitWithRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("<html><body><h1>Success</h1></body></html>")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Test that retry logic is called the expected number of times
	// by creating a new scraper for each test to avoid visited URL cache
	for i := 0; i < 3; i++ {
		cfg := &config.Config{
			Processing: config.ProcessingConfig{
				Concurrency: 1,
				Delay:       0.1,
			},
			Security: config.SecurityConfig{
				RequestTimeout: 5 * time.Second,
			},
			Monitoring: config.MonitoringConfig{
				LogLevel: "error",
			},
		}

		scraper, err := New(cfg)
		require.NoError(t, err)

		// Use a unique path for each attempt
		testURL := server.URL + fmt.Sprintf("/test-retry-path-%d", i)

		// This will trigger server response based on attemptCount
		err = scraper.visitWithRetry(testURL, 1) // Single attempt per scraper
		if i < 2 {
			assert.Error(t, err, "Should fail on attempts 1 and 2")
		} else {
			assert.NoError(t, err, "Should succeed on attempt 3")
		}
	}

	assert.Equal(t, 3, attemptCount, "Should have made 3 total attempts")
}

func TestGetUserAgent(t *testing.T) {
	tests := []struct {
		name     string
		engines  []config.EngineConfig
		expected string
	}{
		{
			name:     "default user agent",
			engines:  []config.EngineConfig{},
			expected: "docs-scraper/1.0",
		},
		{
			name: "custom user agent",
			engines: []config.EngineConfig{
				{Type: "colly", UserAgent: "custom-agent/2.0"},
			},
			expected: "custom-agent/2.0",
		},
		{
			name: "non-colly engine ignored",
			engines: []config.EngineConfig{
				{Type: "chromedp", UserAgent: "browser-agent"},
				{Type: "colly", UserAgent: "colly-agent"},
			},
			expected: "colly-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Engines: tt.engines}
			scraper := &Scraper{config: cfg}
			result := scraper.getUserAgent()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompilePatterns(t *testing.T) {
	tests := []struct {
		name           string
		followPatterns []string
		ignorePatterns []string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "valid patterns",
			followPatterns: []string{"^https://example\\.com/.*"},
			ignorePatterns: []string{".*\\.jpg$"},
			expectError:    false,
		},
		{
			name:           "invalid follow pattern",
			followPatterns: []string{"[invalid regex"},
			expectError:    true,
			errorContains:  "invalid follow pattern",
		},
		{
			name:           "invalid ignore pattern",
			ignorePatterns: []string{"[invalid regex"},
			expectError:    true,
			errorContains:  "invalid ignore pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				FollowPatterns: tt.followPatterns,
				IgnorePatterns: tt.ignorePatterns,
			}
			scraper := &Scraper{config: cfg}

			err := scraper.compilePatterns()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, scraper.followPatterns, len(tt.followPatterns))
				assert.Len(t, scraper.ignorePatterns, len(tt.ignorePatterns))
			}
		})
	}
}

func TestShouldFollow(t *testing.T) {
	tests := []struct {
		name           string
		followPatterns []string
		ignorePatterns []string
		url            string
		expected       bool
	}{
		{
			name:     "privacy URL should be rejected",
			url:      "https://example.com/privacy",
			expected: false,
		},
		{
			name:           "URL matches ignore pattern",
			ignorePatterns: []string{".*\\.jpg$"},
			url:            "https://example.com/image.jpg",
			expected:       false,
		},
		{
			name:           "URL matches follow pattern",
			followPatterns: []string{"^https://example\\.com/docs/.*"},
			url:            "https://example.com/docs/guide",
			expected:       true,
		},
		{
			name:           "URL doesn't match follow pattern",
			followPatterns: []string{"^https://example\\.com/docs/.*"},
			url:            "https://example.com/blog/post",
			expected:       false,
		},
		{
			name:     "no patterns means follow all (except privacy)",
			url:      "https://example.com/anything",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				FollowPatterns: tt.followPatterns,
				IgnorePatterns: tt.ignorePatterns,
			}
			scraper, err := New(cfg)
			require.NoError(t, err)

			result := scraper.shouldFollow(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name             string
		html             string
		titleSelector    string
		expectedContains string
	}{
		{
			name:             "custom selector",
			html:             `<html><head></head><body><h1 class="page-title">Custom Title</h1></body></html>`,
			titleSelector:    ".page-title",
			expectedContains: "Custom Title",
		},
		{
			name:             "default title tag",
			html:             `<html><head><title>Page Title</title></head><body></body></html>`,
			titleSelector:    "",
			expectedContains: "Page Title",
		},
		{
			name:             "fallback to untitled",
			html:             `<html><head></head><body><p>No title</p></body></html>`,
			titleSelector:    "",
			expectedContains: "Untitled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.html))
			}))
			defer server.Close()

			cfg := &config.Config{
				Selectors: config.SelectorConfig{
					Title: tt.titleSelector,
				},
				Processing: config.ProcessingConfig{
					MaxDepth:    1,
					Concurrency: 1,
				},
				Security: config.SecurityConfig{
					RequestTimeout: 5 * time.Second,
				},
				Monitoring: config.MonitoringConfig{
					LogLevel: "error",
				},
			}

			scraper, err := New(cfg)
			require.NoError(t, err)

			// Use reflection or test the handleHTML method indirectly
			// For now, test through integration
			err = scraper.collector.Visit(server.URL)
			assert.NoError(t, err)
		})
	}
}

func TestFindAndFollowLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`
				<html>
				<body>
					<main>
						<p>Homepage content</p>
						<a href="/page1">Page 1</a>
						<a href="/page2">Page 2</a>
						<a href="https://external.com/page">External</a>
						<a href="/privacy">Privacy</a>
					</main>
				</body>
				</html>
			`))
		case "/page1", "/page2":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><body><main><h1>Subpage</h1><p>Subpage content</p></main></body></html>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Processing: config.ProcessingConfig{
			MaxDepth:    2,
			Concurrency: 1,
			Delay:       0.1,
		},
		Security: config.SecurityConfig{
			RequestTimeout:  5 * time.Second,
			ScrapingTimeout: 10 * time.Second,
		},
		Monitoring: config.MonitoringConfig{
			LogLevel: "error",
		},
	}

	scraper, err := New(cfg)
	require.NoError(t, err)

	err = scraper.collector.Visit(server.URL)
	assert.NoError(t, err)

	// Give time for async processing
	scraper.collector.Wait()

	// Check that multiple pages were processed
	pageCount := scraper.aggregator.GetPageCount()
	assert.Greater(t, pageCount, 1, "Should have followed some links")
}

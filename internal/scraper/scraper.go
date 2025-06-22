package scraper

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/sirupsen/logrus"
	"github.com/vladkampov/markdocify/internal/config"
	"github.com/vladkampov/markdocify/internal/converter"
	"github.com/vladkampov/markdocify/internal/aggregator"
	"github.com/vladkampov/markdocify/internal/types"
)

type Scraper struct {
	config    *config.Config
	collector *colly.Collector
	converter *converter.Converter
	aggregator *aggregator.Aggregator
	
	followPatterns []*regexp.Regexp
	ignorePatterns []*regexp.Regexp
	
	visitedURLs sync.Map
	// mu was removed - no longer needed with atomic operations
	pageCount   int64 // Use atomic operations
	
	logger *logrus.Logger
}

const (
	DefaultMaxRetries = 3
	DefaultBackoffBase = 1 * time.Second
	MaxBackoffDelay = 30 * time.Second
)


func New(cfg *config.Config) (*Scraper, error) {
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.Monitoring.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	s := &Scraper{
		config: cfg,
		logger: logger,
	}

	if err := s.compilePatterns(); err != nil {
		return nil, fmt.Errorf("failed to compile patterns: %w", err)
	}

	s.collector = s.createCollector()
	
	converter, err := converter.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create converter: %w", err)
	}
	s.converter = converter

	aggregator, err := aggregator.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create aggregator: %w", err)
	}
	s.aggregator = aggregator

	return s, nil
}

func (s *Scraper) createCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent(s.getUserAgent()),
	)

	// Debug logging is handled in OnRequest callback below

	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: s.config.Processing.Concurrency,
		Delay:       time.Duration(s.config.Processing.Delay * float64(time.Second)),
	}); err != nil {
		s.logger.WithError(err).Warn("Failed to set rate limit")
	}

	c.SetRequestTimeout(s.config.Security.RequestTimeout)

	c.OnRequest(func(r *colly.Request) {
		s.logger.Debugf("Visiting: %s", r.URL.String())
	})

	c.OnHTML("html", s.handleHTML)

	c.OnError(func(r *colly.Response, err error) {
		s.logger.Warnf("Error scraping %s: %v", r.Request.URL, err)
	})

	c.OnResponse(func(r *colly.Response) {
		s.logger.Debugf("Response from %s: %d bytes", r.Request.URL, len(r.Body))
	})

	return c
}

func (s *Scraper) getUserAgent() string {
	for _, engine := range s.config.Engines {
		if engine.Type == "colly" && engine.UserAgent != "" {
			return engine.UserAgent
		}
	}
	return "docs-scraper/1.0"
}

func (s *Scraper) compilePatterns() error {
	for _, pattern := range s.config.FollowPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid follow pattern '%s': %w", pattern, err)
		}
		s.followPatterns = append(s.followPatterns, re)
	}

	for _, pattern := range s.config.IgnorePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid ignore pattern '%s': %w", pattern, err)
		}
		s.ignorePatterns = append(s.ignorePatterns, re)
	}

	return nil
}

func (s *Scraper) shouldFollow(urlStr string) bool {
	// Always skip privacy policy, terms, and legal pages
	if s.isPrivacyOrLegalURL(urlStr) {
		s.logger.Debugf("Skipping privacy/legal URL: %s", urlStr)
		return false
	}

	if len(s.ignorePatterns) > 0 {
		for _, re := range s.ignorePatterns {
			if re.MatchString(urlStr) {
				return false
			}
		}
	}

	if len(s.followPatterns) > 0 {
		for _, re := range s.followPatterns {
			if re.MatchString(urlStr) {
				return true
			}
		}
		return false
	}

	return true
}

func (s *Scraper) isAllowedDomain(urlStr string) bool {
	if len(s.config.Security.AllowedDomains) == 0 {
		return true
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	for _, domain := range s.config.Security.AllowedDomains {
		// Exact match or subdomain match
		if u.Host == domain || strings.HasSuffix(u.Host, "."+domain) {
			return true
		}
	}

	return false
}

func (s *Scraper) isPrivacyOrLegalURL(urlStr string) bool {
	// Common privacy, legal, and non-documentation URLs to skip
	skipPatterns := []string{
		"privacy", "terms", "legal", "cookies", "gdpr",
		"about", "contact", "support", "careers", "jobs",
		"blog(?!/)", "news", "press", "media",
		"pricing", "enterprise", "commercial", "sales",
		"login", "signup", "register", "account", "profile",
		"404", "error", "maintenance", "status",
	}
	
	for _, pattern := range skipPatterns {
		if matched, _ := regexp.MatchString("(?i)/("+pattern+")($|[/?#])", urlStr); matched {
			return true
		}
	}
	return false
}

func (s *Scraper) handleHTML(e *colly.HTMLElement) {
	currentURL := e.Request.URL.String()
	depth := e.Request.Depth
	
	s.logger.WithFields(logrus.Fields{
		"url":   currentURL,
		"depth": depth,
	}).Info("Processing page")
	
	if !s.isAllowedDomain(currentURL) {
		s.logger.WithFields(logrus.Fields{
			"url":    currentURL,
			"reason": "disallowed_domain",
		}).Warn("Skipping page")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"current_depth": depth,
		"max_depth":     s.config.Processing.MaxDepth,
	}).Debug("Depth check")

	if _, visited := s.visitedURLs.LoadOrStore(currentURL, true); visited {
		s.logger.WithFields(logrus.Fields{
			"url":    currentURL,
			"reason": "already_visited",
		}).Debug("Skipping page")
		return
	}

	title := s.extractTitle(e)
	s.logger.WithFields(logrus.Fields{
		"url":   currentURL,
		"title": title,
	}).Debug("Extracted title")
	
	content := s.extractContent(e)

	if content == "" {
		s.logger.WithFields(logrus.Fields{
			"url":    currentURL,
			"reason": "no_content_found",
		}).Warn("Skipping page")
		return
	}
	
	s.logger.WithFields(logrus.Fields{
		"url":            currentURL,
		"content_length": len(content),
		"depth":          depth,
		"title":          title,
	}).Info("Content extracted successfully")

	pageContent := &types.PageContent{
		URL:       currentURL,
		Title:     title,
		Content:   content,
		Depth:     depth,
		Timestamp: time.Now(),
	}

	markdown, err := s.converter.ConvertToMarkdown(pageContent)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"url":   currentURL,
			"error": err.Error(),
		}).Error("Failed to convert to markdown")
		return
	}

	s.aggregator.AddPage(currentURL, title, markdown, depth)

	// Progress reporting for comprehensive scrapes using atomic counter
	currentCount := atomic.AddInt64(&s.pageCount, 1)
	if currentCount%10 == 0 && currentCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"pages_processed": currentCount,
			"milestone":       "progress_report",
		}).Info("📄 Processing milestone reached")
	}

	// Only follow links if we haven't reached max depth
	if depth < s.config.Processing.MaxDepth {
		s.findAndFollowLinks(e)
	} else {
		s.logger.Debugf("Not following links from %s (depth %d >= max %d)", currentURL, depth, s.config.Processing.MaxDepth)
	}
}

func (s *Scraper) extractTitle(e *colly.HTMLElement) string {
	if s.config.Selectors.Title != "" {
		title := e.ChildText(s.config.Selectors.Title)
		if title != "" {
			return s.cleanTitle(strings.TrimSpace(title))
		}
	}
	
	title := e.ChildText("title")
	if title != "" {
		return s.cleanTitle(strings.TrimSpace(title))
	}
	
	return "Untitled"
}

func (s *Scraper) cleanTitle(title string) string {
	// Conservative title cleaning - only remove obvious artifacts
	cleaned := title
	
	// First remove feature status indicators and other long descriptive text
	statusPatterns := []string{
		`\s*This feature is available in the latest.*?React\s*`,
		`\s*This feature is available in the latest Canary\s*`,
		`\s*This feature is available in the latest Experimental version of React\s*`,
	}
	
	for _, pattern := range statusPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		cleaned = re.ReplaceAllString(cleaned, "")
	}
	
	// Then remove site branding patterns (end of title)
	brandingPatterns := []string{
		`\s*–\s*React\s*$`,
		`\s*-\s*React\s*$`,
		`\s*\|\s*React\s*$`,
		`\s*–\s*Stripe\s*$`,
		`\s*-\s*Stripe\s*$`,
		`\s*\|\s*Stripe\s*$`,
		`\s*\|\s*.*Documentation\s*$`,
		`\s*\|\s*.*Docs\s*$`,
	}
	
	for _, pattern := range brandingPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		cleaned = re.ReplaceAllString(cleaned, "")
	}
	
	// Only remove consecutive identical words (conservative deduplication)
	words := strings.Fields(cleaned)
	var deduped []string
	for i, word := range words {
		if i == 0 || !strings.EqualFold(word, words[i-1]) {
			deduped = append(deduped, word)
		}
	}
	cleaned = strings.Join(deduped, " ")
	
	return strings.TrimSpace(cleaned)
}

func (s *Scraper) extractContent(e *colly.HTMLElement) string {
	contentSelector := s.config.Selectors.Content
	if contentSelector == "" {
		contentSelector = "main, article, .content"
	}

	s.logger.Debugf("Using content selector: %s", contentSelector)

	var contentParts []string
	
	e.ForEach(contentSelector, func(i int, el *colly.HTMLElement) {
		s.logger.Debugf("Found content element %d", i)
		
		for _, excludeSelector := range s.config.Selectors.Exclude {
			el.ForEach(excludeSelector, func(j int, excluded *colly.HTMLElement) {
				excluded.DOM.Remove()
			})
		}
		
		html, err := el.DOM.Html()
		if err == nil && strings.TrimSpace(html) != "" {
			s.logger.Debugf("Extracted content length: %d", len(html))
			contentParts = append(contentParts, html)
		}
	})

	result := strings.Join(contentParts, "\n\n")
	s.logger.Debugf("Total extracted content length: %d", len(result))
	return result
}

func (s *Scraper) findAndFollowLinks(e *colly.HTMLElement) {
	e.ForEach("a[href]", func(i int, el *colly.HTMLElement) {
		link := el.Attr("href")
		if link == "" {
			return
		}

		absoluteURL := e.Request.AbsoluteURL(link)
		
		if !s.shouldFollow(absoluteURL) {
			return
		}

		if !s.isAllowedDomain(absoluteURL) {
			return
		}

		if _, visited := s.visitedURLs.Load(absoluteURL); visited {
			return
		}

		s.logger.Debugf("Following link: %s", absoluteURL)
		if err := e.Request.Visit(absoluteURL); err != nil {
			s.logger.WithError(err).Warnf("Failed to visit link: %s", absoluteURL)
		}
	})
}

// Run executes the scraper with default context behavior.
// This is a convenience method that calls RunWithContext with context.Background().
func (s *Scraper) Run() error {
	return s.RunWithContext(context.Background())
}

// RunWithContext executes the scraper with context support for cancellation and timeouts.
// The context can be used to cancel the scraping operation gracefully.
// Returns an error if all start URLs fail, context is cancelled, or output generation fails.
func (s *Scraper) RunWithContext(ctx context.Context) error {
	s.logger.WithFields(logrus.Fields{
		"name":            s.config.Name,
		"output_file":     s.config.OutputFile,
		"start_urls":      len(s.config.StartURLs),
		"max_depth":       s.config.Processing.MaxDepth,
		"scraping_timeout": s.config.Security.ScrapingTimeout.String(),
	}).Info("Starting scraper")

	// Create context with scraping timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.Security.ScrapingTimeout)
	defer cancel()

	done := make(chan error, 1)
	
	go func() {
		defer close(done)
		
		var allErrors []error
		for _, startURL := range s.config.StartURLs {
			select {
			case <-timeoutCtx.Done():
				done <- timeoutCtx.Err()
				return
			default:
				s.logger.WithFields(logrus.Fields{
					"start_url": startURL,
				}).Info("Processing start URL")
				
				if err := s.visitWithRetry(startURL, DefaultMaxRetries); err != nil {
					s.logger.WithFields(logrus.Fields{
						"start_url": startURL,
						"error":     err.Error(),
					}).Error("Failed to visit start URL after retries")
					allErrors = append(allErrors, fmt.Errorf("failed to visit %s: %w", startURL, err))
				}
			}
		}

		// If all start URLs failed, return error
		if len(allErrors) == len(s.config.StartURLs) {
			done <- fmt.Errorf("all start URLs failed: %v", allErrors)
			return
		}

		s.collector.Wait()

		finalPageCount := s.aggregator.GetPageCount()
		s.logger.WithFields(logrus.Fields{
			"total_pages": finalPageCount,
		}).Info("🎉 Scraping completed")
		
		s.logger.Info("📝 Generating comprehensive markdown output...")
		
		if err := s.aggregator.GenerateOutput(); err != nil {
			done <- fmt.Errorf("failed to generate output: %w", err)
			return
		}

		s.logger.WithFields(logrus.Fields{
			"total_pages":   finalPageCount,
			"output_file":   s.config.OutputFile,
			"partial_fails": len(allErrors),
		}).Info("✅ Documentation scraping completed successfully")
		
		// Log any partial failures but don't fail overall if we got some content
		if len(allErrors) > 0 && finalPageCount > 0 {
			s.logger.WithFields(logrus.Fields{
				"failed_urls":   len(allErrors),
				"success_pages": finalPageCount,
			}).Warn("⚠️  Some start URLs failed, but scraping succeeded")
		}
		
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			s.logger.WithFields(logrus.Fields{
				"timeout": s.config.Security.ScrapingTimeout.String(),
				"reason":  "scraping_timeout_exceeded",
			}).Warn("Scraping timed out - consider increasing scraping_timeout in config")
		} else {
			s.logger.WithFields(logrus.Fields{
				"reason": timeoutCtx.Err().Error(),
			}).Warn("Scraping cancelled")
		}
		return timeoutCtx.Err()
	}
}

func (s *Scraper) visitWithRetry(url string, maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := s.collector.Visit(url); err != nil {
			lastErr = err
			if i < maxRetries-1 { // Don't sleep on last attempt
				backoff := time.Duration(math.Pow(2, float64(i))) * DefaultBackoffBase
				if backoff > MaxBackoffDelay {
					backoff = MaxBackoffDelay
				}
				s.logger.WithFields(logrus.Fields{
					"url":         url,
					"retry":       i + 1,
					"max_retries": maxRetries,
					"backoff":     backoff.String(),
					"error":       err.Error(),
				}).Debug("Retrying after error")
				time.Sleep(backoff)
			}
			continue
		}
		if i > 0 {
			s.logger.WithFields(logrus.Fields{
				"url":     url,
				"retries": i,
			}).Info("Successfully recovered after retries")
		}
		return nil
	}
	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
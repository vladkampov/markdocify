package aggregator

import (
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/vladkampov/markdocify/internal/config"
)

const MaxPagesInMemory = 1000

type Aggregator struct {
	config        *config.Config
	pages         []*Page
	mu            sync.RWMutex
	contentHashes map[string]bool
}

type Page struct {
	URL       string
	Title     string
	Content   string
	Depth     int
	Timestamp time.Time
}

func New(cfg *config.Config) (*Aggregator, error) {
	return &Aggregator{
		config:        cfg,
		pages:         make([]*Page, 0),
		contentHashes: make(map[string]bool),
	}, nil
}

func (a *Aggregator) AddPage(url, title, content string, depth int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check for duplicate content using hash
	contentHash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	if a.contentHashes[contentHash] {
		return // Skip duplicate content
	}

	// Memory management warning
	if len(a.pages) >= MaxPagesInMemory {
		// TODO: Implement streaming to temp file for very large sites
		fmt.Printf("Warning: Approaching memory limit with %d pages\n", len(a.pages))
	}

	page := &Page{
		URL:       url,
		Title:     title,
		Content:   content,
		Depth:     depth,
		Timestamp: time.Now(),
	}
	
	a.pages = append(a.pages, page)
	a.contentHashes[contentHash] = true
}

func (a *Aggregator) GetPageCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.pages)
}

func (a *Aggregator) GenerateOutput() error {
	if len(a.pages) == 0 {
		return fmt.Errorf("no pages to aggregate")
	}

	a.sortPages()

	var output strings.Builder

	if a.config.Output.IncludeMetadata {
		a.writeMetadata(&output)
	}

	if a.config.Processing.GenerateTOC {
		a.writeTableOfContents(&output)
	}

	a.writeContent(&output)

	return a.writeToFile(output.String())
}

func (a *Aggregator) sortPages() {
	sort.Slice(a.pages, func(i, j int) bool {
		if a.pages[i].Depth != a.pages[j].Depth {
			return a.pages[i].Depth < a.pages[j].Depth
		}
		return a.pages[i].URL < a.pages[j].URL
	})
}

func (a *Aggregator) writeMetadata(output *strings.Builder) {
	output.WriteString("# " + a.config.Name + "\n\n")
	output.WriteString(fmt.Sprintf("*Generated on %s*\n\n", time.Now().Format("2006-01-02 15:04:05")))
	output.WriteString(fmt.Sprintf("- **Base URL**: %s\n", a.config.BaseURL))
	output.WriteString(fmt.Sprintf("- **Total Pages**: %d\n", len(a.pages)))
	output.WriteString(fmt.Sprintf("- **Max Depth**: %d\n\n", a.config.Processing.MaxDepth))
	output.WriteString("---\n\n")
}

func (a *Aggregator) writeTableOfContents(output *strings.Builder) {
	output.WriteString("## Table of Contents\n\n")
	
	for _, page := range a.pages {
		indent := strings.Repeat("  ", page.Depth)
		anchor := a.createAnchor(page.Title)
		output.WriteString(fmt.Sprintf("%s- [%s](#%s)\n", indent, page.Title, anchor))
	}
	
	output.WriteString("\n---\n\n")
}

func (a *Aggregator) createAnchor(title string) string {
	anchor := strings.ToLower(title)
	anchor = strings.ReplaceAll(anchor, " ", "-")
	anchor = strings.ReplaceAll(anchor, ".", "")
	anchor = strings.ReplaceAll(anchor, "(", "")
	anchor = strings.ReplaceAll(anchor, ")", "")
	anchor = strings.ReplaceAll(anchor, "/", "")
	anchor = strings.ReplaceAll(anchor, "\\", "")
	anchor = strings.ReplaceAll(anchor, ":", "")
	anchor = strings.ReplaceAll(anchor, ";", "")
	anchor = strings.ReplaceAll(anchor, "?", "")
	anchor = strings.ReplaceAll(anchor, "!", "")
	anchor = strings.ReplaceAll(anchor, "@", "")
	anchor = strings.ReplaceAll(anchor, "#", "")
	anchor = strings.ReplaceAll(anchor, "$", "")
	anchor = strings.ReplaceAll(anchor, "%", "")
	anchor = strings.ReplaceAll(anchor, "^", "")
	anchor = strings.ReplaceAll(anchor, "&", "")
	anchor = strings.ReplaceAll(anchor, "*", "")
	anchor = strings.ReplaceAll(anchor, "+", "")
	anchor = strings.ReplaceAll(anchor, "=", "")
	anchor = strings.ReplaceAll(anchor, "[", "")
	anchor = strings.ReplaceAll(anchor, "]", "")
	anchor = strings.ReplaceAll(anchor, "{", "")
	anchor = strings.ReplaceAll(anchor, "}", "")
	anchor = strings.ReplaceAll(anchor, "|", "")
	anchor = strings.ReplaceAll(anchor, "\"", "")
	anchor = strings.ReplaceAll(anchor, "'", "")
	anchor = strings.ReplaceAll(anchor, "<", "")
	anchor = strings.ReplaceAll(anchor, ">", "")
	anchor = strings.ReplaceAll(anchor, ",", "")
	
	for strings.Contains(anchor, "--") {
		anchor = strings.ReplaceAll(anchor, "--", "-")
	}
	
	anchor = strings.Trim(anchor, "-")
	
	return anchor
}

func (a *Aggregator) writeContent(output *strings.Builder) {
	for i, page := range a.pages {
		if i > 0 {
			output.WriteString("\n\n---\n\n")
		}

		pageTitle := page.Title
		if pageTitle == "" || pageTitle == "Untitled" {
			pageTitle = a.extractTitleFromURL(page.URL)
		}

		headingLevel := page.Depth + 1
		if headingLevel > 6 {
			headingLevel = 6
		}
		
		headingPrefix := strings.Repeat("#", headingLevel)
		output.WriteString(fmt.Sprintf("%s %s\n\n", headingPrefix, pageTitle))

		if a.config.Output.IncludeMetadata {
			output.WriteString(fmt.Sprintf("*Source: [%s](%s)*\n\n", page.URL, page.URL))
		}

		content := strings.TrimSpace(page.Content)
		if content != "" {
			output.WriteString(content)
			output.WriteString("\n")
		}
	}
}

func (a *Aggregator) extractTitleFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if lastPart == "" && len(parts) > 1 {
			lastPart = parts[len(parts)-2]
		}
		
		if lastPart != "" {
			title := strings.ReplaceAll(lastPart, "-", " ")
			title = strings.ReplaceAll(title, "_", " ")
			title = titleCase(title)
			return title
		}
	}
	
	return "Untitled"
}

func (a *Aggregator) writeToFile(content string) error {
	file, err := os.Create(a.config.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to output file: %w", err)
	}

	return nil
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
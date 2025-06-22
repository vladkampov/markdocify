package converter

import (
	"fmt"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/vladkampov/markdocify/internal/config"
	"github.com/vladkampov/markdocify/internal/types"
)

type Converter struct {
	config     *config.Config
	sanitizer  *bluemonday.Policy
	mdConverter *md.Converter
}


func New(cfg *config.Config) (*Converter, error) {
	c := &Converter{
		config: cfg,
	}

	c.sanitizer = c.createSanitizer()
	c.mdConverter = c.createMarkdownConverter()

	return c, nil
}

func (c *Converter) createSanitizer() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	
	p.AllowElements("pre", "code", "blockquote", "h1", "h2", "h3", "h4", "h5", "h6")
	p.AllowElements("table", "thead", "tbody", "tr", "th", "td")
	p.AllowElements("ul", "ol", "li", "dl", "dt", "dd")
	p.AllowElements("p", "br", "hr", "div", "span")
	p.AllowElements("strong", "b", "em", "i", "u", "s", "del", "ins")
	p.AllowElements("a").AllowAttrs("href", "title").OnElements("a")
	
	if c.config.Output.PreserveImages {
		p.AllowElements("img").AllowAttrs("src", "alt", "title", "width", "height").OnElements("img")
	}

	p.AllowAttrs("class").OnElements("pre", "code")
	
	if !c.config.Output.InlineStyles {
		p.AllowAttrs("style").OnElements("*")
	}

	return p
}

func (c *Converter) createMarkdownConverter() *md.Converter {
	converter := md.NewConverter("", true, nil)
	
	converter.Use(plugin.GitHubFlavored())
	
	return converter
}

func (c *Converter) ConvertToMarkdown(page *types.PageContent) (string, error) {
	if page.Content == "" {
		return "", fmt.Errorf("no content to convert")
	}

	var content string = page.Content

	if c.config.Processing.SanitizeHTML {
		content = c.sanitizer.Sanitize(content)
	}

	markdown, err := c.mdConverter.ConvertString(content)
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to markdown: %w", err)
	}

	markdown = c.postProcessMarkdown(markdown)

	if c.config.Output.IncludeMetadata {
		metadata := c.generateMetadata(page)
		markdown = metadata + "\n\n" + markdown
	}

	return markdown, nil
}

func (c *Converter) postProcessMarkdown(markdown string) string {
	lines := strings.Split(markdown, "\n")
	var processedLines []string

	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		
		if strings.TrimSpace(line) == "" {
			if len(processedLines) == 0 || processedLines[len(processedLines)-1] != "" {
				processedLines = append(processedLines, "")
			}
		} else {
			processedLines = append(processedLines, line)
		}
	}

	for len(processedLines) > 0 && processedLines[len(processedLines)-1] == "" {
		processedLines = processedLines[:len(processedLines)-1]
	}

	return strings.Join(processedLines, "\n")
}

func (c *Converter) generateMetadata(page *types.PageContent) string {
	var metadata []string
	
	metadata = append(metadata, fmt.Sprintf("<!-- Source: %s -->", page.URL))
	metadata = append(metadata, fmt.Sprintf("<!-- Title: %s -->", page.Title))
	metadata = append(metadata, fmt.Sprintf("<!-- Depth: %d -->", page.Depth))
	
	return strings.Join(metadata, "\n")
}
// Package types contains shared data structures used across the application.
package types

import "time"

// PageContent represents the scraped content of a single web page.
type PageContent struct {
	URL       string
	Title     string
	Content   string
	Depth     int
	Timestamp time.Time
}

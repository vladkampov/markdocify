package types

import "time"

type PageContent struct {
	URL       string
	Title     string
	Content   string
	Depth     int
	Timestamp time.Time
}

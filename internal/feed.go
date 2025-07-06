package internal

import "time"

type FeedResponse struct {
	Rss       RSS2JSONResponse
	UpdatedAt time.Time
}

// Top-level response structure
type RSS2JSONResponse struct {
	Status string `json:"status"`
	Feed   Feed   `json:"feed"`
	Items  []Item `json:"items"`
}

// Feed structure
type Feed struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

// Item structure (for each RSS/blog post)
type Item struct {
	Title       string   `json:"title"`
	PubDate     string   `json:"pubDate"`
	Link        string   `json:"link"`
	GUID        string   `json:"guid"`
	Author      string   `json:"author"`
	Thumbnail   string   `json:"thumbnail"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Enclosure   struct{} `json:"enclosure"`
	Categories  []string `json:"categories"`
}

// ParsedPubDate parses the PubDate string to time.Time
func (i *Item) ParsedPubDate() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", i.PubDate)
}

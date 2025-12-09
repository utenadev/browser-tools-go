package models

// SearchResult represents a single search engine result.
type SearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
	Content string `json:"content,omitempty"`
}

// HnSubmission represents a single Hacker News submission.
type HnSubmission struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Points   int    `json:"points"`
	Author   string `json:"author"`
	Time     string `json:"time"`
	Comments int    `json:"comments"`
	HnURL    string `json:"hnUrl"`
}

// ElementInfo represents extracted information from a DOM element.
type ElementInfo struct {
	Tag      string                 `json:"tag"`
	Text     string                 `json:"text"`
	Attrs    map[string]string      `json:"attrs"`
	Rect     map[string]interface{} `json:"rect"`
	Children []ElementInfo          `json:"children"`
}

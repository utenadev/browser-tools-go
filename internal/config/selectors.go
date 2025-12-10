package config

// SelectorConfig holds CSS selectors for scraping various websites.
type SelectorConfig struct {
	Google GoogleSelectors `json:"google"`
	HN     HNSelectors     `json:"hn"`
}

// GoogleSelectors holds CSS selectors for Google search results.
type GoogleSelectors struct {
	// ResultContainer is the selector for the search results container
	ResultContainer string `json:"resultContainer"`
	// Title is the selector for result titles
	Title string `json:"title"`
	// TitleFallback is a fallback selector for result titles
	TitleFallback string `json:"titleFallback"`
	// Link is the selector for result links
	Link string `json:"link"`
	// Snippet is the selector for result snippets
	Snippet string `json:"snippet"`
}

// HNSelectors holds CSS selectors for Hacker News.
type HNSelectors struct {
	// Container is the selector for the stories container
	Container string `json:"container"`
	// Title is the selector for story titles
	Title string `json:"title"`
	// URL is the selector for story URLs
	URL string `json:"url"`
	// Score is the selector for story scores
	Score string `json:"score"`
	// Author is the selector for story authors
	Author string `json:"author"`
	// Time is the selector for story timestamps
	Time string `json:"time"`
	// Comments is the selector for comment links
	Comments string `json:"comments"`
}

// DefaultSelectors returns the default CSS selectors.
func DefaultSelectors() *SelectorConfig {
	return &SelectorConfig{
		Google: GoogleSelectors{
			ResultContainer: "div#search",
			Title:           "h3",
			TitleFallback:   "a h3",
			Link:            "div#search a",
			Snippet:         "div.VwiC3b",
		},
		HN: HNSelectors{
			Container: "table.itemlist",
			Title:     "span.titleline > a",
			URL:       "span.titleline > a",
			Score:     ".score",
			Author:    ".hnuser",
			Time:      "span.age a",
			Comments:  "td.subtext > a",
		},
	}
}

// global selector config, initialized with defaults
var selectors = DefaultSelectors()

// GetSelectors returns the current selector configuration.
func GetSelectors() *SelectorConfig {
	return selectors
}

// SetSelectors allows overriding the selector configuration (useful for testing).
func SetSelectors(s *SelectorConfig) {
	selectors = s
}

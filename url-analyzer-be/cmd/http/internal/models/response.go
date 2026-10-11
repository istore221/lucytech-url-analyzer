package models

type Link struct {
	URL string `json:"url"`
}

type PageAnalysis struct {
	URL               string         `json:"url"`
	HTMLVersion       string         `json:"html_version"`
	Title             string         `json:"title"`
	Headings          map[string]int `json:"headings"`
	InternalLinks     []Link         `json:"internal_links"`
	ExternalLinks     []Link         `json:"external_links"`
	InaccessibleLinks []Link         `json:"inaccessible_links"`
	HasLoginForm      bool           `json:"has_login_form"`
}

type TargetURL struct {
	URL string `json:"url"`
}

type ScrapeRequest struct {
	URLs []TargetURL `json:"urls"`
}

type ScrapeError struct {
	Code    string   `json:"code"` // ex: SCRAPE_ERROR, SCRAPE_TIMEOUT
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

type ScrapeItemResult struct {
	URL      string        `json:"url"`
	Analysis *PageAnalysis `json:"analysis,omitempty"`
	Error    *ScrapeError  `json:"error,omitempty"`
}

type APIError struct {
	Code    string   `json:"code"`              // e.g., "INVALID_PAYLOAD", "INTERNAL_SERVER_ERROR"
	Message string   `json:"message"`           // High-level human-readable message
	Details []string `json:"details,omitempty"` // Specific field or binding validation issues
}

type APIResponse struct {
	Status int       `json:"status"`
	Data   any       `json:"data,omitempty"`
	Error  *APIError `json:"error,omitempty"`
}

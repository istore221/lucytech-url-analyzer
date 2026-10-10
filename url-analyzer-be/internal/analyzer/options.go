package analyzer

import (
	"net/http"
	"strings"
)

type Option func(*Analyzer)

func WithWorkers(n int) Option {
	return func(a *Analyzer) {
		if n > 0 {
			a.workers = n
		}
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(a *Analyzer) {
		if client != nil {
			a.client = client
		}
	}
}

func WithIgnoredSchemes(schemes ...string) Option {
	return func(a *Analyzer) {
		for _, s := range schemes {
			a.ignoredSchemes[strings.ToLower(strings.TrimSpace(s))] = true
		}
	}
}

func WithScraper(scraper Scraper) Option {
	return func(a *Analyzer) {
		if scraper != nil {
			a.scraper = scraper
		}
	}
}

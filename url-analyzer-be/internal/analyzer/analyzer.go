package analyzer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultWorkers = 10
	defaultTimeout = 15 * time.Second
	maxBodySize    = 2 * 1024 * 1024
)

type Analyzer struct {
	client         *http.Client
	scraper        Scraper
	workers        int
	ignoredSchemes map[string]bool
}

func New(opts ...Option) *Analyzer {
	var t *http.Transport
	if dt, ok := http.DefaultTransport.(*http.Transport); ok {
		t = dt.Clone()
	} else {
		t = &http.Transport{}
	}
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100

	a := &Analyzer{
		workers: defaultWorkers,
		scraper: &DefaultHTMLScraper{},
		client: &http.Client{
			Timeout:   defaultTimeout,
			Transport: t,
		},
		ignoredSchemes: map[string]bool{
			"javascript": true,
			"vbscript":   true,
			"data":       true,
			"blob":       true,

			"mailto": true,
			"tel":    true,
			"sms":    true,
			"callto": true,

			"file":   true,
			"intent": true,
		},
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

func (a *Analyzer) Analyze(ctx context.Context, targetURL string) (*PageAnalysis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}
	parsedTarget, err := url.ParseRequestURI(targetURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedTarget.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &HTTPError{
			StatusCode:  0,
			Description: "Failed to establish connection to the server.",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, &HTTPError{
			StatusCode:  resp.StatusCode,
			Description: http.StatusText(resp.StatusCode),
		}
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	limitReader := io.LimitReader(resp.Body, maxBodySize)
	scrapeData, err := a.scraper.Scrape(ScrapeRequest{
		Reader: limitReader,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScrape, err)
	}

	_, _ = io.Copy(io.Discard, resp.Body) // drain for conn reuse

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var internalLinks, externalLinks, allValidatableLinks []string
	for _, rawLink := range scrapeData.RawLinks {
		resolvedURL, resolveErr := parsedTarget.Parse(strings.TrimSpace(rawLink))
		if resolveErr != nil || a.ignoredSchemes[strings.ToLower(resolvedURL.Scheme)] {
			continue
		}

		linkStr := resolvedURL.String()
		allValidatableLinks = append(allValidatableLinks, linkStr)

		if resolvedURL.Host == parsedTarget.Host || resolvedURL.Host == "" {
			internalLinks = append(internalLinks, linkStr)
		} else {
			externalLinks = append(externalLinks, linkStr)
		}
	}

	inaccessibleLinks, err := a.getInaccessibleLinks(ctx, allValidatableLinks)
	if err != nil {
		return nil, err
	}

	return &PageAnalysis{
		URL:               parsedTarget.String(),
		HTMLVersion:       scrapeData.HTMLVersion,
		Title:             scrapeData.Title,
		Headings:          scrapeData.Headings,
		InternalLinks:     internalLinks,
		ExternalLinks:     externalLinks,
		InaccessibleLinks: inaccessibleLinks,
		HasLoginForm:      scrapeData.HasLoginForm,
	}, nil
}

func (a *Analyzer) getInaccessibleLinks(ctx context.Context, links []string) ([]string, error) {
	if len(links) == 0 {
		return nil, nil
	}

	jobs := make(chan string, len(links))
	results := make(chan string, len(links))
	var wg sync.WaitGroup

	numWorkers := a.workers
	if len(links) < numWorkers {
		numWorkers = len(links)
	}

	// Start Workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for link := range jobs {

				if ctx.Err() != nil {
					return
				}

				if a.isLinkInaccessible(ctx, link) {
					select {
					case <-ctx.Done():
						return
					case results <- link:
					}
				}
			}
		}()
	}

	for _, link := range links {
		jobs <- link
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var inaccessible []string
	for brokenLink := range results {
		inaccessible = append(inaccessible, brokenLink)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return inaccessible, nil
}

func (a *Analyzer) isLinkInaccessible(ctx context.Context, link string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return true
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return true
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 400
}

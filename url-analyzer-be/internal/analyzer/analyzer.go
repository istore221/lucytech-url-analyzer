package analyzer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	defaultWorkers = 10
	defaultTimeout = 15 * time.Second
	maxBodySize    = 2 * 1024 * 1024
)

type CategorizedLinks struct {
	Internal    []Link
	External    []Link
	Validatable []Link
}

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

			"android-app": true,
			"ios-app":     true,

			"file":   true,
			"ftp":    true,
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
		Reader:    limitReader,
		TargetURL: parsedTarget,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScrape, err)
	}

	_, _ = io.Copy(io.Discard, resp.Body) // drain for conn reuse

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	links, inaccessibleLinks, err := a.processAndValidateLinks(ctx, parsedTarget, scrapeData.RawLinks)
	if err != nil {
		return nil, err
	}

	return &PageAnalysis{
		URL:               parsedTarget.String(),
		HTMLVersion:       scrapeData.HTMLVersion,
		Title:             scrapeData.Title,
		Headings:          scrapeData.Headings,
		InternalLinks:     links.Internal,
		ExternalLinks:     links.External,
		InaccessibleLinks: inaccessibleLinks,
		HasLoginForm:      scrapeData.HasLoginForm,
	}, nil
}

func (a *Analyzer) processAndValidateLinks(ctx context.Context, baseURL *url.URL, rawLinks []Link) (CategorizedLinks, []Link, error) {
	links := a.categorizeLinks(baseURL, rawLinks)

	inaccessibleLinks, err := a.getInaccessibleLinks(ctx, links.Validatable)
	if err != nil {
		return CategorizedLinks{}, nil, err
	}

	return links, inaccessibleLinks, nil
}

func (a *Analyzer) categorizeLinks(baseURL *url.URL, rawLinks []Link) CategorizedLinks {

	categorized := CategorizedLinks{
		Validatable: make([]Link, 0, len(rawLinks)),
		Internal:    make([]Link, 0, len(rawLinks)/2),
		External:    make([]Link, 0, len(rawLinks)/2),
	}

	seen := make(map[string]struct{}, len(rawLinks))

	for _, rawLink := range rawLinks {
		rawURLStr := strings.TrimSpace(rawLink.URL())
		if rawURLStr == "" {
			continue
		}

		resolvedURL, err := baseURL.Parse(rawURLStr)
		if err != nil {
			continue
		}

		scheme := strings.ToLower(resolvedURL.Scheme)
		if a.ignoredSchemes[scheme] {
			continue
		}

		normalizedURL := a.NormalizeURL(resolvedURL)
		normalizedStr := normalizedURL.String()

		if _, exists := seen[normalizedStr]; exists {
			continue
		}
		seen[normalizedStr] = struct{}{}

		scrapedLink := ScrapedLink(normalizedStr)
		categorized.Validatable = append(categorized.Validatable, scrapedLink)

		if normalizedURL.Host == baseURL.Host || normalizedURL.Host == "" {
			categorized.Internal = append(categorized.Internal, scrapedLink)
		} else {
			categorized.External = append(categorized.External, scrapedLink)
		}
	}

	return categorized
}

func (a *Analyzer) NormalizeURL(u *url.URL) *url.URL {
	cloned := *u

	cloned.Scheme = strings.ToLower(cloned.Scheme)
	cloned.Host = strings.ToLower(cloned.Host)

	port := cloned.Port()
	if (cloned.Scheme == "http" && port == "80") ||
		(cloned.Scheme == "https" && port == "443") {
		cloned.Host = cloned.Hostname()
	}

	cloned.Fragment = ""
	cloned.RawFragment = ""

	if cloned.RawQuery != "" {
		cloned.RawQuery = cloned.Query().Encode()
	}

	if cloned.Path != "" {

		cloned.Path = path.Clean(cloned.Path)

		if cloned.Path == "/" {
			cloned.Path = ""
		}
	}

	return &cloned
}

func (a *Analyzer) getInaccessibleLinks(ctx context.Context, links []Link) ([]Link, error) {
	if len(links) == 0 {
		return nil, nil
	}

	jobs := make(chan Link, len(links))
	results := make(chan Link, len(links))
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

	var inaccessible []Link
	for brokenLink := range results {
		inaccessible = append(inaccessible, brokenLink)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return inaccessible, nil
}

func (a *Analyzer) isLinkInaccessible(ctx context.Context, link Link) bool {

	inaccessible, err := a.checkURL(ctx, http.MethodHead, link.URL())
	if err == nil && !inaccessible {
		return false
	}

	inaccessible, err = a.checkURL(ctx, http.MethodGet, link.URL())
	if err != nil {
		return true
	}

	return inaccessible
}

func (a *Analyzer) checkURL(ctx context.Context, method, targetURL string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, method, targetURL, nil)
	if err != nil {
		return true, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := a.client.Do(req)
	if err != nil {
		return true, err
	}
	defer resp.Body.Close()

	// If HEAD is blocked or disallowed (403, 405, 501), treat as inconclusive to trigger GET fallback
	if method == http.MethodHead {
		if resp.StatusCode == http.StatusMethodNotAllowed ||
			resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == http.StatusNotImplemented {
			return true, fmt.Errorf("HEAD method not supported or blocked (status %d)", resp.StatusCode)
		}
	}

	return resp.StatusCode >= 400, nil // ignoring 301,302 redirects for now
}

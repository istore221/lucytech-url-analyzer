package analyzer_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/istore221/lucytech-url-analyzer/internal/analyzer"
	"github.com/istore221/lucytech-url-analyzer/internal/analyzer/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/h2non/gock.v1"
)

func TestAnalyzer_Analyze(t *testing.T) {
	targetURL := "https://example.com"

	t.Run("it returns success for page analysis and detects inaccessible links", func(t *testing.T) {
		defer gock.Off()

		gock.New(targetURL).
			Get("/").
			Reply(http.StatusOK)

		gock.New(targetURL).
			Head("/internal").
			Reply(http.StatusOK)

		gock.New(targetURL).
			Get("/internal").
			MatchParam("a", "test").
			MatchParam("b", "test").
			Reply(http.StatusOK)

		gock.New(targetURL).
			Get("/internal").
			MatchParam("a", "test").
			MatchParam("b", "test").
			MatchParam("c", "test").
			Reply(http.StatusOK)

		gock.New("https://external.com").
			Get("/").
			Reply(http.StatusOK)

		gock.New("https://broken-external.com").
			Get("/").
			Reply(http.StatusNotFound)

		gock.New("https://network-error.com").
			Get("/").
			ReplyError(errors.New("connection reset by peer"))

		mockScraper := mocks.NewMockScraper(t)
		expectedData := &analyzer.ScrapeData{
			HTMLVersion: "HTML5",
			Title:       "Example Page",
			Headings:    map[string]int{"h1": 1},
			RawLinks: []analyzer.Link{
				analyzer.ScrapedLink("/internal"),
				analyzer.ScrapedLink("   https://external.com    "),
				analyzer.ScrapedLink("https://network-error.com"),
				analyzer.ScrapedLink("https://broken-external.com"),
				analyzer.ScrapedLink("/internal?b=test&a=test#section"),
				analyzer.ScrapedLink("/internal?a=test&b=test"),
				analyzer.ScrapedLink("/internal?c=test&b=test&a=test"),
				analyzer.ScrapedLink("javascript:alert(1)"),
				analyzer.ScrapedLink("DATA:image/png;base64,123"),
				analyzer.ScrapedLink("DATA:image/png;base64,123"),
				analyzer.ScrapedLink("tel:+1234567890"),
				analyzer.ScrapedLink("ftp://://example.com"),
			},
			HasLoginForm: false,
		}

		mockScraper.On("Scrape", mock.AnythingOfType("analyzer.ScrapeRequest")).
			Return(expectedData, nil)

		client := &http.Client{Transport: &http.Transport{}}
		gock.InterceptClient(client)

		a := analyzer.New(
			analyzer.WithWorkers(1),
			analyzer.WithScraper(mockScraper),
			analyzer.WithHTTPClient(client),
		)

		res, err := a.Analyze(context.Background(), targetURL)

		assert.NoError(t, err)
		assert.Equal(t, "https://example.com", res.URL)
		assert.Equal(t, "HTML5", res.HTMLVersion)
		assert.Equal(t, "Example Page", res.Title)

		assert.ElementsMatch(t, []analyzer.Link{
			analyzer.ScrapedLink("https://example.com/internal"),
			analyzer.ScrapedLink("https://example.com/internal?a=test&b=test"),
			analyzer.ScrapedLink("https://example.com/internal?a=test&b=test&c=test"),
		}, res.InternalLinks)

		assert.ElementsMatch(t, []analyzer.Link{
			analyzer.ScrapedLink("https://external.com"),
			analyzer.ScrapedLink("https://broken-external.com"),
			analyzer.ScrapedLink("https://network-error.com"),
		}, res.ExternalLinks)

		assert.ElementsMatch(t, []analyzer.Link{
			analyzer.ScrapedLink("https://broken-external.com"),
			analyzer.ScrapedLink("https://network-error.com"),
		}, res.InaccessibleLinks)
	})

	t.Run("it returns success for page analysis and detects with no links in the page", func(t *testing.T) {
		defer gock.Off()

		mockScraper := mocks.NewMockScraper(t)
		expectedData := &analyzer.ScrapeData{
			HTMLVersion:  "HTML5",
			Title:        "Example Page",
			Headings:     map[string]int{"h1": 1},
			RawLinks:     []analyzer.Link{},
			HasLoginForm: false,
		}

		mockScraper.On("Scrape", mock.AnythingOfType("analyzer.ScrapeRequest")).
			Return(expectedData, nil)

		client := &http.Client{Transport: &http.Transport{}}
		gock.InterceptClient(client)

		a := analyzer.New(
			analyzer.WithWorkers(1),
			analyzer.WithScraper(mockScraper),
			analyzer.WithHTTPClient(client),
		)

		res, err := a.Analyze(context.Background(), targetURL)

		assert.NoError(t, err)
		assert.Equal(t, "https://example.com", res.URL)
		assert.Equal(t, "HTML5", res.HTMLVersion)
		assert.Equal(t, "Example Page", res.Title)

		assert.Empty(t, res.InternalLinks)
		assert.Empty(t, res.ExternalLinks)
		assert.Empty(t, res.InaccessibleLinks)
	})

	t.Run("automatically prefixes target URL without scheme with https", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://noscheme.com").
			Get("/").
			Reply(http.StatusOK)

		mockScraper := mocks.NewMockScraper(t)
		mockScraper.On("Scrape", mock.AnythingOfType("analyzer.ScrapeRequest")).
			Return(&analyzer.ScrapeData{}, nil)

		client := &http.Client{Transport: &http.Transport{}}
		gock.InterceptClient(client)

		a := analyzer.New(
			analyzer.WithWorkers(1),
			analyzer.WithScraper(mockScraper),
			analyzer.WithHTTPClient(client),
		)

		res, err := a.Analyze(context.Background(), "noscheme.com")

		assert.NoError(t, err)
		assert.Equal(t, "https://noscheme.com", res.URL)
	})

	t.Run("returns error on invalid URL", func(t *testing.T) {
		a := analyzer.New()
		res, err := a.Analyze(context.Background(), "http://192.168.0.%31/")

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, analyzer.ErrInvalidURL)
	})

	t.Run("returns error when initial HTTP request fails", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://unreachable.com").
			Get("/").
			ReplyError(errors.New("dial tcp: lookup failed"))

		client := &http.Client{Transport: &http.Transport{}}
		gock.InterceptClient(client)

		a := analyzer.New(
			analyzer.WithWorkers(1),
			analyzer.WithHTTPClient(client),
		)

		res, err := a.Analyze(context.Background(), "https://unreachable.com")

		assert.Error(t, err)
		assert.Nil(t, res)

		var httpErr *analyzer.HTTPError
		assert.True(t, errors.As(err, &httpErr))
		assert.Equal(t, 0, httpErr.StatusCode)
	})

	t.Run("returns HTTPError on non-2xx status code", func(t *testing.T) {
		defer gock.Off()

		gock.New(targetURL).
			Get("/").
			Reply(http.StatusInternalServerError)

		client := &http.Client{Transport: &http.Transport{}}
		gock.InterceptClient(client)

		a := analyzer.New(
			analyzer.WithWorkers(1),
			analyzer.WithHTTPClient(client),
		)

		res, err := a.Analyze(context.Background(), targetURL)

		assert.Error(t, err)
		assert.Nil(t, res)

		var httpErr *analyzer.HTTPError
		assert.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusInternalServerError, httpErr.StatusCode)
	})

	t.Run("returns error when scraper fails", func(t *testing.T) {
		defer gock.Off()

		gock.New(targetURL).
			Get("/").
			Reply(http.StatusOK)

		scrapeErr := errors.New("malformed html")
		mockScraper := mocks.NewMockScraper(t)
		mockScraper.On("Scrape", mock.AnythingOfType("analyzer.ScrapeRequest")).
			Return(nil, scrapeErr)

		client := &http.Client{Transport: &http.Transport{}}
		gock.InterceptClient(client)

		a := analyzer.New(
			analyzer.WithWorkers(1),
			analyzer.WithScraper(mockScraper),
			analyzer.WithHTTPClient(client),
		)

		res, err := a.Analyze(context.Background(), targetURL)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, analyzer.ErrScrape)
	})

	t.Run("returns context error when canceled prior to execution", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		a := analyzer.New(
			analyzer.WithWorkers(1),
		)
		res, err := a.Analyze(ctx, targetURL)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("ignores unsupported URL schemes", func(t *testing.T) {
		defer gock.Off()

		gock.New(targetURL).
			Get("/").
			Reply(http.StatusOK)

		mockScraper := mocks.NewMockScraper(t)
		mockScraper.On("Scrape", mock.AnythingOfType("analyzer.ScrapeRequest")).
			Return(&analyzer.ScrapeData{
				RawLinks: []analyzer.Link{
					analyzer.ScrapedLink("mailto:user@example.com"),
					analyzer.ScrapedLink("tel:+1234567890"),
					analyzer.ScrapedLink("javascript:void(0)"),
					analyzer.ScrapedLink("file:///etc/passwd"),
				},
			}, nil)

		client := &http.Client{Transport: &http.Transport{}}
		gock.InterceptClient(client)

		a := analyzer.New(
			analyzer.WithScraper(mockScraper),
			analyzer.WithHTTPClient(client),
		)

		res, err := a.Analyze(context.Background(), targetURL)

		assert.NoError(t, err)
		assert.Empty(t, res.InternalLinks)
		assert.Empty(t, res.ExternalLinks)
		assert.Empty(t, res.InaccessibleLinks)
	})

	t.Run("returns context error when canceled during worker processing", func(t *testing.T) {
		defer gock.Off()

		gock.New(targetURL).
			Get("/").
			Reply(http.StatusOK)

		mockScraper := mocks.NewMockScraper(t)
		mockScraper.On("Scrape", mock.AnythingOfType("analyzer.ScrapeRequest")).
			Return(&analyzer.ScrapeData{
				RawLinks: []analyzer.Link{
					analyzer.ScrapedLink("/slow-link-1"),
					analyzer.ScrapedLink("/slow-link-2"),
				},
			}, nil)

		client := &http.Client{Transport: &http.Transport{}}
		gock.InterceptClient(client)

		a := analyzer.New(
			analyzer.WithScraper(mockScraper),
			analyzer.WithHTTPClient(client),
			analyzer.WithWorkers(1),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		gock.New(targetURL).
			Get("/slow-link-1").
			Map(func(r *http.Request) *http.Request {
				time.Sleep(50 * time.Millisecond)
				return r
			}).
			Reply(http.StatusOK)

		res, err := a.Analyze(ctx, targetURL)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

}

func TestNormalizeURL(t *testing.T) {

	a := &analyzer.Analyzer{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase scheme and host",
			input:    "HTTPS://EXAMPLE.COM/page",
			expected: "https://example.com/page",
		},
		{
			name:     "remove default http port",
			input:    "http://example.com:80/path",
			expected: "http://example.com/path",
		},
		{
			name:     "remove default https port",
			input:    "https://example.com:443/path",
			expected: "https://example.com/path",
		},
		{
			name:     "keep non-default port",
			input:    "https://example.com:8443/path",
			expected: "https://example.com:8443/path",
		},
		{
			name:     "remove fragment",
			input:    "https://example.com/page#section-1",
			expected: "https://example.com/page",
		},
		{
			name:     "sort query parameters",
			input:    "https://example.com/api?z=1&a=2&b=3",
			expected: "https://example.com/api?a=2&b=3&z=1",
		},
		{
			name:     "remove root trailing slash",
			input:    "https://example.com/",
			expected: "https://example.com",
		},
		{
			name:     "remove path trailing slash",
			input:    "https://example.com/about/",
			expected: "https://example.com/about",
		},
		{
			name:     "clean duplicate slashes",
			input:    "https://example.com//api///v1//users",
			expected: "https://example.com/api/v1/users",
		},
		{
			name:     "resolve relative paths (.. and .)",
			input:    "https://example.com/a/b/../c/./d",
			expected: "https://example.com/a/c/d",
		},
		{
			name:     "complex combination of all rules",
			input:    "HTTPS://EXAMPLE.COM:443//api/v1/../v2/?z=99&a=1#top",
			expected: "https://example.com/api/v2?a=1&z=99",
		},
		{
			name:     "root trailing slash with query params",
			input:    "https://example.com/?query=1",
			expected: "https://example.com?query=1",
		},
		{
			name:     "already clean url remains unchanged",
			input:    "https://example.com/clean/path?q=search",
			expected: "https://example.com/clean/path?q=search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("failed to parse input url: %v", err)
			}

			normalized := a.NormalizeURL(parsedURL)
			actual := normalized.String()

			if actual != tt.expected {
				t.Errorf("\nexpected: %s\ngot:      %s", tt.expected, actual)
			}
		})
	}
}

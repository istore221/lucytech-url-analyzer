package analyzer_test

import (
	"context"
	"errors"
	"net/http"
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
			Get("/internal").
			Reply(http.StatusOK)

		gock.New(targetURL).
			Get("/internal").
			MatchParam("arg", "test").
			Reply(http.StatusOK)

		gock.New("https://external.com").
			Get("/").
			Reply(http.StatusOK)

		gock.New(targetURL).
			Get("/broken").
			Reply(http.StatusNotFound)

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
				analyzer.ScrapedLink("/broken"),
				analyzer.ScrapedLink("https://broken-external.com"),
				analyzer.ScrapedLink("/internal?arg=test#section"),
				analyzer.ScrapedLink("javascript:alert(1)"),
				analyzer.ScrapedLink("DATA:image/png;base64,123"),
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

		assert.ElementsMatch(t, []string{
			"https://example.com/internal",
			"https://example.com/broken",
			"https://example.com/internal?arg=test#section",
		}, res.InternalLinks)

		assert.ElementsMatch(t, []string{
			"https://external.com",
			"https://broken-external.com",
			"https://network-error.com",
		}, res.ExternalLinks)

		assert.ElementsMatch(t, []string{
			"https://example.com/broken",
			"https://broken-external.com",
			"https://network-error.com",
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

	t.Run("treats HTTP 405 Method Not Allowed as accessible", func(t *testing.T) {
		defer gock.Off()

		gock.New(targetURL).
			Get("/").
			Reply(http.StatusOK)

		gock.New(targetURL).
			Get("/method-not-allowed").
			Reply(http.StatusMethodNotAllowed)

		mockScraper := mocks.NewMockScraper(t)
		expectedData := &analyzer.ScrapeData{
			HTMLVersion: "HTML5",
			Title:       "Example Page",
			Headings:    map[string]int{"h1": 1},
			RawLinks: []analyzer.Link{
				analyzer.ScrapedLink("/method-not-allowed"),
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
		assert.ElementsMatch(t, []string{
			"https://example.com/method-not-allowed",
		}, res.InternalLinks)

		assert.Empty(t, res.InaccessibleLinks)
	})
}

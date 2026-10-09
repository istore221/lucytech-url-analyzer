package analyzer_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	analyzer "github.com/istore221/lucytech-url-analyzer"
	"github.com/istore221/lucytech-url-analyzer/mocks"
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
			RawLinks: []string{
				"/internal",
				"https://external.com",
				"#section",
				"https://network-error.com",
				"/broken",
				"https://broken-external.com",
				"/internal?arg=test#section",
			},
			HasLoginForm: false,
		}

		mockScraper.On("Scrape", mock.AnythingOfType("analyzer.ScrapeRequest")).
			Return(expectedData, nil)

		client := &http.Client{Transport: &http.Transport{}}
		gock.InterceptClient(client)

		a := analyzer.New(
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
			"https://example.com#section",
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
}

package scraper

import (
	"context"
	"errors"
	"testing"

	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/models"
	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/scraper/mocks"
	"github.com/istore221/lucytech-url-analyzer/internal/analyzer"
	"github.com/stretchr/testify/assert"
)

type mockLink struct {
	url string
}

func (m mockLink) URL() string {
	return m.url
}

func TestNewService(t *testing.T) {
	mockEngine := mocks.NewMockScraperEngine(t)
	service := NewService(mockEngine)

	assert.NotNil(t, service)
	assert.Equal(t, mockEngine, service.engine)
}

func TestAnalyzeService_Analyze_Success(t *testing.T) {
	mockEngine := mocks.NewMockScraperEngine(t)
	service := NewService(mockEngine)

	ctx := context.Background()
	targetURL := "https://example.com"

	analyzerResult := &analyzer.PageAnalysis{
		URL:         targetURL,
		HTMLVersion: "HTML5",
		Title:       "Example Title",
		Headings: map[string]int{
			"h1": 1,
			"h2": 2,
		},
		InternalLinks: []analyzer.Link{
			mockLink{url: "https://example.com/about"},
		},
		ExternalLinks: []analyzer.Link{
			mockLink{url: "https://google.com"},
		},
		InaccessibleLinks: []analyzer.Link{
			mockLink{url: "https://example.com/404"},
		},
		HasLoginForm: true,
	}

	mockEngine.On("Analyze", ctx, targetURL).Return(analyzerResult, nil)

	result, err := service.Analyze(ctx, targetURL)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, targetURL, result.URL)
	assert.Equal(t, "HTML5", result.HTMLVersion)
	assert.Equal(t, "Example Title", result.Title)
	assert.Equal(t, map[string]int{"h1": 1, "h2": 2}, result.Headings)
	assert.Equal(t, []models.Link{{URL: "https://example.com/about"}}, result.InternalLinks)
	assert.Equal(t, []models.Link{{URL: "https://google.com"}}, result.ExternalLinks)
	assert.Equal(t, []models.Link{{URL: "https://example.com/404"}}, result.InaccessibleLinks)
	assert.True(t, result.HasLoginForm)
}

func TestAnalyzeService_Analyze_EngineError(t *testing.T) {
	mockEngine := mocks.NewMockScraperEngine(t)
	service := NewService(mockEngine)

	ctx := context.Background()
	targetURL := "https://example.com"
	expectedErr := errors.New("network error")

	mockEngine.On("Analyze", ctx, targetURL).Return((*analyzer.PageAnalysis)(nil), expectedErr)

	result, err := service.Analyze(ctx, targetURL)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
}

func TestAnalyzeService_Analyze_NilAnalysis(t *testing.T) {
	mockEngine := mocks.NewMockScraperEngine(t)
	service := NewService(mockEngine)

	ctx := context.Background()
	targetURL := "https://example.com"

	mockEngine.On("Analyze", ctx, targetURL).Return((*analyzer.PageAnalysis)(nil), nil)

	result, err := service.Analyze(ctx, targetURL)

	assert.Error(t, err)
	assert.EqualError(t, err, "analyzer returned empty result")
	assert.Nil(t, result)
}

func TestAnalyzeService_Analyze_NilLinksMapToEmptySlices(t *testing.T) {
	mockEngine := mocks.NewMockScraperEngine(t)
	service := NewService(mockEngine)

	ctx := context.Background()
	targetURL := "https://example.com"

	analyzerResult := &analyzer.PageAnalysis{
		URL:               targetURL,
		InternalLinks:     nil,
		ExternalLinks:     nil,
		InaccessibleLinks: nil,
	}

	mockEngine.On("Analyze", ctx, targetURL).Return(analyzerResult, nil)

	result, err := service.Analyze(ctx, targetURL)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.InternalLinks)
	assert.Empty(t, result.InternalLinks)
	assert.NotNil(t, result.ExternalLinks)
	assert.Empty(t, result.ExternalLinks)
	assert.NotNil(t, result.InaccessibleLinks)
	assert.Empty(t, result.InaccessibleLinks)
}

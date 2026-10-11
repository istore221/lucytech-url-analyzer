package scraper

import (
	"context"
	"errors"

	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/models"
	"github.com/istore221/lucytech-url-analyzer/internal/analyzer"
)

type Engine interface {
	Analyze(ctx context.Context, targetURL string) (*models.PageAnalysis, error)
}

type ScraperEngine interface {
	Analyze(ctx context.Context, targetURL string) (*analyzer.PageAnalysis, error)
}

type AnalyzeService struct {
	engine ScraperEngine
}

func NewService(engine ScraperEngine) *AnalyzeService {
	return &AnalyzeService{
		engine: engine,
	}
}

func (s *AnalyzeService) Analyze(ctx context.Context, targetURL string) (*models.PageAnalysis, error) {

	analysis, err := s.engine.Analyze(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	if analysis == nil {
		return nil, errors.New("analyzer returned empty result")
	}

	return mapToAPIModel(analysis), nil
}

func mapToAPIModel(src *analyzer.PageAnalysis) *models.PageAnalysis {
	return &models.PageAnalysis{
		URL:               src.URL,
		HTMLVersion:       src.HTMLVersion,
		Title:             src.Title,
		Headings:          src.Headings,
		InternalLinks:     mapLinks(src.InternalLinks),
		ExternalLinks:     mapLinks(src.ExternalLinks),
		InaccessibleLinks: mapLinks(src.InaccessibleLinks),
		HasLoginForm:      src.HasLoginForm,
	}
}

func mapLinks(srcLinks []analyzer.Link) []models.Link {
	if srcLinks == nil {
		return []models.Link{}
	}

	links := make([]models.Link, len(srcLinks))
	for i, l := range srcLinks {
		links[i] = models.Link{
			URL: l.URL(),
		}
	}
	return links
}

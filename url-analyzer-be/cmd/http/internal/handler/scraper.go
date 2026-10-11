package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/apperrors"
	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/models"
	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/scraper"
)

type ScrapeHandler struct {
	engine                  scraper.Engine
	scrapeTimeout           time.Duration
	MaxHTTPBatchConcurrency int
	MaxHTTPBatchSize        int
}

func NewScrapeHandler(engine scraper.Engine,
	timeout time.Duration,
	MaxHTTPBatchConcurrency,
	MaxHTTPBatchSize int) *ScrapeHandler {

	return &ScrapeHandler{
		engine:                  engine,
		scrapeTimeout:           timeout,
		MaxHTTPBatchConcurrency: MaxHTTPBatchConcurrency,
		MaxHTTPBatchSize:        MaxHTTPBatchSize,
	}
}

func (h *ScrapeHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Status: http.StatusOK,
		Data: gin.H{
			"status": "UP",
		},
	})
}

// validateURL checks for presence, valid structure, and HTTP/HTTPS scheme.
func validateURL(target string) (string, error) {
	if target == "" {
		return "Missing value", errors.New("url field is required")
	}
	parsed, err := url.ParseRequestURI(target)
	if err != nil {
		return "URL parsing error", errors.New("malformed URL format")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "Unsupported protocol", errors.New("URL scheme must be http or https")
	}
	if parsed.Host == "" {
		return "Missing hostname", errors.New("URL missing host")
	}
	return "", nil
}

func (h *ScrapeHandler) Scrape(c *gin.Context) {
	var req models.ScrapeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status: http.StatusBadRequest,
			Error:  apperrors.NewAPIError(apperrors.CodeInvalidJSON, apperrors.MsgInvalidJSON, err.Error()),
		})
		return
	}

	if len(req.URLs) == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status: http.StatusBadRequest,
			Error:  apperrors.NewAPIError(apperrors.CodeEmptyBatch, apperrors.MsgEmptyBatch),
		})
		return
	}

	if len(req.URLs) > h.MaxHTTPBatchSize {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status: http.StatusBadRequest,
			Error: apperrors.NewAPIError(
				apperrors.CodeBatchSizeExceeded,
				apperrors.MsgBatchSizeExceeded,
				fmt.Sprintf("Batch size %d exceeds maximum limit of %d URLs", len(req.URLs), h.MaxHTTPBatchSize),
			),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.scrapeTimeout)
	defer cancel()

	results := make([]models.ScrapeItemResult, len(req.URLs))
	var wg sync.WaitGroup

	limitConcurrency := make(chan struct{}, h.MaxHTTPBatchConcurrency)

	for i, target := range req.URLs {
		limitConcurrency <- struct{}{}
		wg.Add(1)

		go func(index int, item models.TargetURL) {
			defer wg.Done()
			defer func() { <-limitConcurrency }()

			if detail, err := validateURL(item.URL); err != nil {
				results[index] = models.ScrapeItemResult{
					URL:   item.URL,
					Error: apperrors.NewScrapeError(apperrors.CodeInvalidURL, err.Error(), detail),
				}
				return
			}

			analysis, err := h.engine.Analyze(ctx, item.URL)
			if err != nil {
				errCode := apperrors.CodeScrapeFailed
				errMsg := "Failed to complete webpage analysis"
				detail := err.Error()

				if errors.Is(err, context.DeadlineExceeded) {
					errCode = apperrors.CodeTimeout
					errMsg = apperrors.MsgTimeout
					detail = "Execution exceeded configured timeout duration"
				}

				results[index] = models.ScrapeItemResult{
					URL:   item.URL,
					Error: apperrors.NewScrapeError(errCode, errMsg, detail),
				}
				return
			}

			results[index] = models.ScrapeItemResult{
				URL:      item.URL,
				Analysis: analysis,
			}
		}(i, target)
	}

	wg.Wait()

	c.JSON(http.StatusOK, models.APIResponse{
		Status: http.StatusOK,
		Data:   results,
	})
}

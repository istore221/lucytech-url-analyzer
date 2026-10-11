package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/apperrors"
	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/handler/mocks"
	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewScrapeHandler(t *testing.T) {
	mockEngine := mocks.NewMockEngine(t)
	timeout := 5 * time.Second
	concurrency := 2
	batchSize := 10

	h := NewScrapeHandler(mockEngine, timeout, concurrency, batchSize)

	assert.NotNil(t, h)
	assert.Equal(t, mockEngine, h.engine)
	assert.Equal(t, timeout, h.scrapeTimeout)
	assert.Equal(t, concurrency, h.MaxHTTPBatchConcurrency)
	assert.Equal(t, batchSize, h.MaxHTTPBatchSize)
}

func TestScrapeHandler_Health(t *testing.T) {
	router := setupTestRouter()
	mockEngine := mocks.NewMockEngine(t)
	h := NewScrapeHandler(mockEngine, 5*time.Second, 2, 10)

	router.GET("/health", h.Health)

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	expectedJSON := `{
		"status": 200,
		"data": {
			"status": "UP"
		}
	}`
	assert.JSONEq(t, expectedJSON, w.Body.String())
}

func TestScrapeHandler_Scrape_InvalidJSON(t *testing.T) {
	router := setupTestRouter()
	mockEngine := mocks.NewMockEngine(t)
	h := NewScrapeHandler(mockEngine, 5*time.Second, 2, 10)

	router.POST("/scrape", h.Scrape)

	reqBody := `invalid json`

	req, err := http.NewRequest(http.MethodPost, "/scrape", bytes.NewBufferString(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), fmt.Sprintf(`"code":"%s"`, apperrors.CodeInvalidJSON))
}

func TestScrapeHandler_Scrape_EmptyBatch(t *testing.T) {
	router := setupTestRouter()
	mockEngine := mocks.NewMockEngine(t)
	h := NewScrapeHandler(mockEngine, 5*time.Second, 2, 10)

	router.POST("/scrape", h.Scrape)

	reqBody := `{
		"urls": []
	}`

	req, err := http.NewRequest(http.MethodPost, "/scrape", bytes.NewBufferString(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	expectedJSON := fmt.Sprintf(`{
		"status": 400,
		"error": {
			"code": "%s",
			"message": "%s"
		}
	}`, apperrors.CodeEmptyBatch, apperrors.MsgEmptyBatch)

	assert.JSONEq(t, expectedJSON, w.Body.String())
}

func TestScrapeHandler_Scrape_BatchSizeExceeded(t *testing.T) {
	router := setupTestRouter()
	mockEngine := mocks.NewMockEngine(t)
	h := NewScrapeHandler(mockEngine, 5*time.Second, 2, 1)

	router.POST("/scrape", h.Scrape)

	reqBody := `{
		"urls": [
			{"url": "https://example.com/1"},
			{"url": "https://example.com/2"}
		]
	}`

	req, err := http.NewRequest(http.MethodPost, "/scrape", bytes.NewBufferString(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	expectedJSON := fmt.Sprintf(`{
		"status": 400,
		"error": {
			"code": "%s",
			"message": "%s",
			"details": [
				"Batch size 2 exceeds maximum limit of 1 URLs"
			]
		}
	}`, apperrors.CodeBatchSizeExceeded, apperrors.MsgBatchSizeExceeded)

	assert.JSONEq(t, expectedJSON, w.Body.String())
}

func TestScrapeHandler_Scrape_Success(t *testing.T) {
	router := setupTestRouter()
	mockEngine := mocks.NewMockEngine(t)
	h := NewScrapeHandler(mockEngine, 5*time.Second, 2, 10)

	router.POST("/scrape", h.Scrape)

	target := "https://www.scrapethissite.com/login/"
	mockAnalysis := &models.PageAnalysis{
		URL:          target,
		HTMLVersion:  "HTML5",
		Title:        "Login Page",
		Headings:     map[string]int{"h1": 1},
		HasLoginForm: true,
	}

	mockEngine.On("Analyze", mock.Anything, target).Return(mockAnalysis, nil)

	reqBody := `{
		"urls": [
			{"url": "https://www.scrapethissite.com/login/"}
		]
	}`

	req, err := http.NewRequest(http.MethodPost, "/scrape", bytes.NewBufferString(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	expectedJSON := `{
		"status": 200,
		"data": [
			{
				"url": "https://www.scrapethissite.com/login/",
				"analysis": {
					"url": "https://www.scrapethissite.com/login/",
					"html_version": "HTML5",
					"title": "Login Page",
					"headings": {
						"h1": 1
					},
					"internal_links": null,
					"external_links": null,
					"inaccessible_links": null,
					"has_login_form": true
				}
			}
		]
	}`
	assert.JSONEq(t, expectedJSON, w.Body.String())
}

func TestScrapeHandler_Scrape_ValidationErrors(t *testing.T) {
	tests := []struct {
		name         string
		reqBody      string
		expectedJSON string
	}{
		{
			name: "missing value",
			reqBody: `{
				"urls": [{"url": ""}]
			}`,
			expectedJSON: fmt.Sprintf(`{
				"status": 200,
				"data": [
					{
						"url": "",
						"error": {
							"code": "%s",
							"message": "url field is required",
							"details": ["Missing value"]
						}
					}
				]
			}`, apperrors.CodeInvalidURL),
		},
		{
			name: "unsupported protocol",
			reqBody: `{
				"urls": [{"url": "ftp://example.com"}]
			}`,
			expectedJSON: fmt.Sprintf(`{
				"status": 200,
				"data": [
					{
						"url": "ftp://example.com",
						"error": {
							"code": "%s",
							"message": "URL scheme must be http or https",
							"details": ["Unsupported protocol"]
						}
					}
				]
			}`, apperrors.CodeInvalidURL),
		},
		{
			name: "missing hostname",
			reqBody: `{
				"urls": [{"url": "http://"}]
			}`,
			expectedJSON: fmt.Sprintf(`{
				"status": 200,
				"data": [
					{
						"url": "http://",
						"error": {
							"code": "%s",
							"message": "URL missing host",
							"details": ["Missing hostname"]
						}
					}
				]
			}`, apperrors.CodeInvalidURL),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			mockEngine := mocks.NewMockEngine(t)
			h := NewScrapeHandler(mockEngine, 5*time.Second, 2, 10)

			router.POST("/scrape", h.Scrape)

			req, err := http.NewRequest(http.MethodPost, "/scrape", bytes.NewBufferString(tt.reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}

func TestScrapeHandler_Scrape_EngineErrors(t *testing.T) {
	tests := []struct {
		name         string
		engineErr    error
		reqBody      string
		expectedJSON string
	}{
		{
			name:      "timeout failure",
			engineErr: context.DeadlineExceeded,
			reqBody: `{
				"urls": [{"url": "https://example.com"}]
			}`,
			expectedJSON: fmt.Sprintf(`{
				"status": 200,
				"data": [
					{
						"url": "https://example.com",
						"error": {
							"code": "%s",
							"message": "%s",
							"details": ["Execution exceeded configured timeout duration"]
						}
					}
				]
			}`, apperrors.CodeTimeout, apperrors.MsgTimeout),
		},
		{
			name:      "generic scrape failure",
			engineErr: errors.New("connection reset"),
			reqBody: `{
				"urls": [{"url": "https://example.com"}]
			}`,
			expectedJSON: fmt.Sprintf(`{
				"status": 200,
				"data": [
					{
						"url": "https://example.com",
						"error": {
							"code": "%s",
							"message": "Failed to complete webpage analysis",
							"details": ["connection reset"]
						}
					}
				]
			}`, apperrors.CodeScrapeFailed),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			mockEngine := mocks.NewMockEngine(t)
			h := NewScrapeHandler(mockEngine, 5*time.Second, 2, 10)

			router.POST("/scrape", h.Scrape)

			target := "https://example.com"
			mockEngine.On("Analyze", mock.Anything, target).Return((*models.PageAnalysis)(nil), tt.engineErr)

			req, err := http.NewRequest(http.MethodPost, "/scrape", bytes.NewBufferString(tt.reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.JSONEq(t, tt.expectedJSON, w.Body.String())
		})
	}
}

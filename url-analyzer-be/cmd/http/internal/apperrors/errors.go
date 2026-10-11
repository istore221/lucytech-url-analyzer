package apperrors

import "github.com/istore221/lucytech-url-analyzer/cmd/http/internal/models"

// Top-Level HTTP Error Codes
const (
	CodeInvalidJSON       = "INVALID_JSON"
	CodeEmptyBatch        = "EMPTY_BATCH"
	CodeInternalErr       = "INTERNAL_SERVER_ERROR"
	CodeBatchSizeExceeded = "BATCH_SIZE_EXCEEDED"
)

// Item-Level Scrape Error Codes
const (
	CodeInvalidURL   = "INVALID_URL"
	CodeScrapeFailed = "SCRAPE_FAILED"
	CodeTimeout      = "TIMEOUT"
)

// Standard Error Messages
const (
	MsgInvalidJSON       = "Malformed JSON payload in request body"
	MsgEmptyBatch        = "The 'urls' array must contain at least one item"
	MsgTimeout           = "Analysis timed out"
	MsgBatchSizeExceeded = "Batch payload exceeds maximum allowed size"
)

func NewAPIError(code, message string, details ...string) *models.APIError {
	return &models.APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func NewScrapeError(code, message string, details ...string) *models.ScrapeError {
	return &models.ScrapeError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

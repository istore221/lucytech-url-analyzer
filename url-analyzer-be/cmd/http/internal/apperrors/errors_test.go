package apperrors

import (
	"testing"

	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewAPIError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		details  []string
		expected *models.APIError
	}{
		{
			name:    "without details",
			code:    CodeInvalidJSON,
			message: MsgInvalidJSON,
			details: nil,
			expected: &models.APIError{
				Code:    CodeInvalidJSON,
				Message: MsgInvalidJSON,
				Details: nil,
			},
		},
		{
			name:    "with single detail",
			code:    CodeEmptyBatch,
			message: MsgEmptyBatch,
			details: []string{"batch cannot be empty"},
			expected: &models.APIError{
				Code:    CodeEmptyBatch,
				Message: MsgEmptyBatch,
				Details: []string{"batch cannot be empty"},
			},
		},
		{
			name:    "with multiple details",
			code:    CodeBatchSizeExceeded,
			message: MsgBatchSizeExceeded,
			details: []string{"Batch size 20 exceeds limit of 10", "Payload size 5MB exceeds max limit"},
			expected: &models.APIError{
				Code:    CodeBatchSizeExceeded,
				Message: MsgBatchSizeExceeded,
				Details: []string{"Batch size 20 exceeds limit of 10", "Payload size 5MB exceeds max limit"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.code, tt.message, tt.details...)

			assert.NotNil(t, err)
			assert.Equal(t, tt.expected.Code, err.Code)
			assert.Equal(t, tt.expected.Message, err.Message)
			assert.Equal(t, tt.expected.Details, err.Details)
		})
	}
}

func TestNewScrapeError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		details  []string
		expected *models.ScrapeError
	}{
		{
			name:    "without details",
			code:    CodeTimeout,
			message: MsgTimeout,
			details: nil,
			expected: &models.ScrapeError{
				Code:    CodeTimeout,
				Message: MsgTimeout,
				Details: nil,
			},
		},
		{
			name:    "with single detail",
			code:    CodeInvalidURL,
			message: "URL missing host",
			details: []string{"Missing hostname"},
			expected: &models.ScrapeError{
				Code:    CodeInvalidURL,
				Message: "URL missing host",
				Details: []string{"Missing hostname"},
			},
		},
		{
			name:    "with multiple details",
			code:    CodeScrapeFailed,
			message: "Failed to complete webpage analysis",
			details: []string{"connection reset", "DNS lookup failed"},
			expected: &models.ScrapeError{
				Code:    CodeScrapeFailed,
				Message: "Failed to complete webpage analysis",
				Details: []string{"connection reset", "DNS lookup failed"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewScrapeError(tt.code, tt.message, tt.details...)

			assert.NotNil(t, err)
			assert.Equal(t, tt.expected.Code, err.Code)
			assert.Equal(t, tt.expected.Message, err.Message)
			assert.Equal(t, tt.expected.Details, err.Details)
		})
	}
}

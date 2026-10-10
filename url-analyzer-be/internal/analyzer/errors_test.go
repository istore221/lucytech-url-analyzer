package analyzer_test

import (
	"errors"
	"testing"

	"github.com/istore221/lucytech-url-analyzer/internal/analyzer"
)

func TestErrorsExist(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrRequestAborted",
			err:      analyzer.ErrRequestAborted,
			expected: "request aborted by context",
		},
		{
			name:     "ErrNetwork",
			err:      analyzer.ErrNetwork,
			expected: "network error",
		},
		{
			name:     "ErrInvalidURL",
			err:      analyzer.ErrInvalidURL,
			expected: "invalid URL",
		},
		{
			name:     "ErrScrape",
			err:      analyzer.ErrScrape,
			expected: "scraping failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("expected error variable to be initialized, got nil")
			}
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHTTPError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      analyzer.HTTPError
		expected string
	}{
		{
			name: "Status code 0 (unreachable)",
			err: analyzer.HTTPError{
				StatusCode:  0,
				Description: "connection refused",
			},
			expected: "unreachable: connection refused",
		},
		{
			name: "Status code 404 (not found)",
			err: analyzer.HTTPError{
				StatusCode:  404,
				Description: "Not Found",
			},
			expected: "HTTP 404: Not Found",
		},
		{
			name: "Status code 500 (internal server error)",
			err: analyzer.HTTPError{
				StatusCode:  500,
				Description: "Internal Server Error",
			},
			expected: "HTTP 500: Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHTTPError_AsErrorInterface(t *testing.T) {

	var err error = &analyzer.HTTPError{
		StatusCode:  403,
		Description: "Forbidden",
	}

	if got := err.Error(); got != "HTTP 403: Forbidden" {
		t.Errorf("got %q, want %q", got, "HTTP 403: Forbidden")
	}

	var httpErr *analyzer.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected errors.As to match *analyzer.HTTPError")
	}
	if httpErr.StatusCode != 403 {
		t.Errorf("expected StatusCode 403, got %d", httpErr.StatusCode)
	}
}

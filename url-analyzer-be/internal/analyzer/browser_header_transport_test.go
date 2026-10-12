package analyzer

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockRoundTripper struct {
	capturedReq *http.Request
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.capturedReq = req

	return &http.Response{StatusCode: http.StatusOK}, nil
}

func TestBrowserHeaderTransport(t *testing.T) {

	mockBase := &mockRoundTripper{}
	transport := &browserHeaderTransport{
		base: mockBase,
	}

	origReq, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	assert.NoError(t, err, "Failed to create request")

	resp, err := transport.RoundTrip(origReq)
	assert.NoError(t, err, "RoundTrip returned an unexpected error")
	assert.NotNil(t, resp, "Response should not be nil")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK")

	assert.NotNil(t, mockBase.capturedReq, "The underlying transport did not receive a request")
	injectedReq := mockBase.capturedReq

	assert.Equal(t, browserUserAgent, injectedReq.Header.Get("User-Agent"))
	assert.Equal(t, browserAccept, injectedReq.Header.Get("Accept"))
	assert.Equal(t, browserLanguage, injectedReq.Header.Get("Accept-Language"))

	assert.Empty(t, origReq.Header.Get("User-Agent"), "Original request mutated: User-Agent")
	assert.Empty(t, origReq.Header.Get("Accept"), "Original request mutated: Accept")
	assert.Empty(t, origReq.Header.Get("Accept-Language"), "Original request mutated: Accept-Language")
}

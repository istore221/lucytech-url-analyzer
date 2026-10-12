package analyzer

import (
	"net/http"
)

const (
	browserUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	browserAccept    = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"
	browserLanguage  = "en-US,en;q=0.9"
)

type browserHeaderTransport struct {
	base http.RoundTripper
}

func (t *browserHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())

	clonedReq.Header.Set("User-Agent", browserUserAgent)
	clonedReq.Header.Set("Accept", browserAccept)
	clonedReq.Header.Set("Accept-Language", browserLanguage)

	return t.base.RoundTrip(clonedReq)
}

package analyzer

import (
	"errors"
	"fmt"
)

var (

	// ErrInvalidURL indicates the provided string is not a valid URL.
	ErrInvalidURL = errors.New("invalid URL")

	// ErrScrape indicates parsing the HTML document failed.
	ErrScrape = errors.New("scraping failed")
)

type HTTPError struct {
	StatusCode  int
	Description string
}

func (e *HTTPError) Error() string {
	if e.StatusCode == 0 {
		return fmt.Sprintf("unreachable: %s", e.Description)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Description)
}

package util

import (
	"fmt"
	"net/url"
)

func NormalizeURL(rawURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", rawURL, err)
	}
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""
	return parsedURL, nil
}

package strava

import (
	"fmt"
	"net/http"
	"strings"

	"go.mattglei.ch/lcp/internal/api"
)

func sendStravaAPIRequest[T any](client *http.Client, path string, tokens Tokens) (T, error) {
	var zero T

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://www.strava.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		return zero, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokens.Access)

	resp, err := api.RequestJSON[T](logPrefix, client, req)
	if err != nil {
		return zero, fmt.Errorf("making strava api request: %w", err)
	}
	return resp, nil
}

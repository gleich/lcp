package strava

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.mattglei.ch/lcp-2/internal/apis"
	"go.mattglei.ch/timber"
)

func sendStravaAPIRequest[T any](client *http.Client, path string, tokens Tokens) (T, error) {
	var zeroValue T

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://www.strava.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		return zeroValue, fmt.Errorf("%w failed to create request", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokens.Access)

	resp, err := apis.SendRequest[T](client, req)
	if err != nil {
		if !errors.Is(err, apis.IgnoreError) {
			timber.Error(err, "failed to make strava API request")
		}
		return zeroValue, err
	}
	return resp, nil
}

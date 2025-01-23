package strava

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"pkg.mattglei.ch/lcp-2/internal/apis"
	"pkg.mattglei.ch/timber"
)

func sendStravaAPIRequest[T any](client *http.Client, path string, tokens tokens) (T, error) {
	var zeroValue T

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://www.strava.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		timber.Error(err, "failed to create request")
		return zeroValue, err
	}
	req.Header.Set("Authorization", "Bearer "+tokens.Access)

	resp, err := apis.SendRequest[T](client, req)
	if err != nil {
		if !errors.Is(err, apis.WarningError) {
			timber.Error(err, "failed to make strava API request")
		}
		return zeroValue, err
	}
	return resp, nil
}

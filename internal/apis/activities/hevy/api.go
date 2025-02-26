package hevy

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.mattglei.ch/lcp-2/internal/apis"
	"go.mattglei.ch/lcp-2/internal/secrets"
)

func sendHevyAPIRequest[T any](client *http.Client, path string) (T, error) {
	var zeroValue T
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://api.hevyapp.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		return zeroValue, fmt.Errorf("%w failed to create request", err)
	}
	req.Header.Set("api-key", secrets.ENV.HevyAccessToken)

	resp, err := apis.SendRequest[T](client, req)
	if err != nil {
		if !errors.Is(err, apis.IgnoreError) {
			return zeroValue, fmt.Errorf("%w failed to make hevy API request", err)
		}
		return zeroValue, err
	}
	return resp, nil
}

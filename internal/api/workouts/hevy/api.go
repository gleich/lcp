package hevy

import (
	"fmt"
	"net/http"
	"strings"

	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/secrets"
)

func sendHevyAPIRequest[T any](client *http.Client, path string) (T, error) {
	var zeroValue T
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://api.hevyapp.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		return zeroValue, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("api-key", secrets.ENV.HevyAccessToken)

	resp, err := api.RequestJSON[T]("[hevy]", client, req)
	if err != nil {
		return zeroValue, fmt.Errorf("making hevy api request: %w", err)
	}
	return resp, nil
}

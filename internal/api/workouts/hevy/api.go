package hevy

import (
	"fmt"
	"net/http"
	"strings"

	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/tlog"
)

func sendHevyAPIRequest[T any](task tlog.Task, client *http.Client, path string) (T, error) {
	var zero T
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://api.hevyapp.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		return zero, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("api-key", secrets.ENV.HevyAccessToken)

	resp, err := api.RequestJSON[T](task, client, req)
	if err != nil {
		return zero, fmt.Errorf("making hevy api request: %w", err)
	}
	return resp, nil
}

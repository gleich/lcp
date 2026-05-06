package jellyfin

import (
	"fmt"
	"net/http"
	"strings"

	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/secrets"
)

func sendJellyfinAPIRequest[T any](client *http.Client, path string) (T, error) {
	var zero T
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://gleich.tv/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		return zero, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Emby-Token", secrets.ENV.JellyfinKey)

	resp, err := api.RequestJSON[T](client, req, logAttr)
	if err != nil {
		return zero, fmt.Errorf("making jellyfin API request: %w", err)
	}
	return resp, nil
}

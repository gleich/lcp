package applemusic

import (
	"fmt"
	"net/http"
	"strings"

	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/secrets"
)

func sendAppleMusicAPIRequest[T any](client *http.Client, path string) (T, error) {
	var zeroValue T
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://api.music.apple.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		return zeroValue, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+secrets.ENV.AppleMusicAppToken)
	req.Header.Set("Music-User-Token", secrets.ENV.AppleMusicUserToken)

	resp, err := api.RequestJSON[T](cacheInstance.LogPrefix(), client, req)
	if err != nil {
		return zeroValue, fmt.Errorf("making apple music API request: %w", err)
	}
	return resp, nil
}

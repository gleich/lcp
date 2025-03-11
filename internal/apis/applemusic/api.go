package applemusic

import (
	"fmt"
	"net/http"
	"strings"

	"go.mattglei.ch/lcp-2/internal/apis"
	"go.mattglei.ch/lcp-2/internal/secrets"
)

func sendAppleMusicAPIRequest[T any](client *http.Client, path string) (T, error) {
	var zeroValue T
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://api.music.apple.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		return zeroValue, fmt.Errorf("%w failed to create request", err)
	}
	req.Header.Set("Authorization", "Bearer "+secrets.ENV.AppleMusicAppToken)
	req.Header.Set("Music-User-Token", secrets.ENV.AppleMusicUserToken)

	resp, err := apis.RequestJSON[T](logPrefix, client, req)
	if err != nil {
		return zeroValue, fmt.Errorf("%w failed to make apple music API request", err)
	}
	return resp, nil
}

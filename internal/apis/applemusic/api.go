package applemusic

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"pkg.mattglei.ch/lcp-2/internal/apis"
	"pkg.mattglei.ch/lcp-2/internal/secrets"
)

func sendAppleMusicAPIRequest[T any](client *http.Client, path string) (T, error) {
	var zeroValue T
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://api.music.apple.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		return zeroValue, fmt.Errorf("%v failed to create request", err)
	}
	req.Header.Set("Authorization", "Bearer "+secrets.ENV.AppleMusicAppToken)
	req.Header.Set("Music-User-Token", secrets.ENV.AppleMusicUserToken)

	resp, err := apis.SendRequest[T](client, req)
	if err != nil {
		if !errors.Is(err, apis.IgnoreError) {
			return zeroValue, fmt.Errorf("%v failed to make apple music API request", err)
		}
		return zeroValue, err
	}
	return resp, nil
}

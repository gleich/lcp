package applemusic

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"pkg.mattglei.ch/lcp-2/internal/apis"
	"pkg.mattglei.ch/lcp-2/internal/secrets"
	"pkg.mattglei.ch/timber"
)

func sendAppleMusicAPIRequest[T any](path string) (T, error) {
	var zeroValue T
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://api.music.apple.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		timber.Error(err, "failed to create request")
		return zeroValue, err
	}
	req.Header.Set("Authorization", "Bearer "+secrets.ENV.AppleMusicAppToken)
	req.Header.Set("Music-User-Token", secrets.ENV.AppleMusicUserToken)

	resp, err := apis.SendRequest[T](req)
	if err != nil {
		if !errors.Is(err, apis.WarningError) {
			timber.Error(err, "failed to make apple music API request")
		}
		return zeroValue, err
	}
	return resp, nil
}

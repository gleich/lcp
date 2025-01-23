package strava

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"pkg.mattglei.ch/lcp-2/internal/apis"
	"pkg.mattglei.ch/lcp-2/internal/secrets"
	"pkg.mattglei.ch/timber"
)

type tokens struct {
	Access    string `json:"access_token"`
	Refresh   string `json:"refresh_token"`
	ExpiresAt int64  `json:"expires_at"`
}

func loadTokens() tokens {
	return tokens{
		Access:    secrets.ENV.StravaAccessToken,
		Refresh:   secrets.ENV.StravaRefreshToken,
		ExpiresAt: 0, // starts at zero to force a refresh on boot
	}
}

func (t *tokens) refreshIfNeeded(client *http.Client) {
	// subtract 60 to ensure that token doesn't expire in the next 60 seconds
	if t.ExpiresAt-60 >= time.Now().Unix() {
		return
	}

	params := url.Values{
		"client_id":     {secrets.ENV.StravaClientID},
		"client_secret": {secrets.ENV.StravaClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {t.Refresh},
		"code":          {secrets.ENV.StravaOAuthCode},
	}
	req, err := http.NewRequest(
		http.MethodPost,
		"https://www.strava.com/oauth/token?"+params.Encode(),
		nil,
	)
	if err != nil {
		timber.Error(err, "creating request for new token failed")
		return
	}

	tokens, err := apis.SendRequest[tokens](client, req)
	if err != nil {
		if !errors.Is(err, apis.IgnoreError) {
			timber.Error(err, "failed to refresh tokens")
		}
		return
	}

	*t = tokens
	timber.Done("loaded new strava access token:", t.Access)
}

package strava

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.mattglei.ch/lcp-2/internal/apis"
	"go.mattglei.ch/lcp-2/internal/secrets"
	"go.mattglei.ch/timber"
)

type Tokens struct {
	Access    string `json:"access_token"`
	Refresh   string `json:"refresh_token"`
	ExpiresAt int64  `json:"expires_at"`
}

func LoadTokens() Tokens {
	return Tokens{
		Access:    secrets.ENV.StravaAccessToken,
		Refresh:   secrets.ENV.StravaRefreshToken,
		ExpiresAt: 0, // starts at zero to force a refresh on boot
	}
}

func (t *Tokens) RefreshIfNeeded(client *http.Client) error {
	// subtract 60 to ensure that token doesn't expire in the next 60 seconds
	if t.ExpiresAt-60 >= time.Now().Unix() {
		return nil
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
		return fmt.Errorf("%w creating request for new token failed", err)
	}

	tokens, err := apis.SendRequest[Tokens](client, req)
	if err != nil {
		if !errors.Is(err, apis.IgnoreError) {
			return fmt.Errorf("%w failed to fetch refresh tokens", err)
		}
		return err
	}

	*t = tokens
	timber.Done(logPrefix, "new access token", t.Access)
	return nil
}

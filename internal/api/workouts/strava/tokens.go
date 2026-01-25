package strava

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/timber"
)

type Tokens struct {
	Access    string `json:"access_token"`
	Refresh   string `json:"refresh_token"`
	ExpiresAt int64  `json:"expires_at"`
}

func LoadTokens() Tokens {
	return Tokens{
		Access:    "",
		Refresh:   secrets.ENV.StravaRefreshToken,
		ExpiresAt: 0, // starts at zero to force a refresh on boot
	}
}

func (t *Tokens) RefreshIfExpired(client *http.Client) error {
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
		return fmt.Errorf("creating request for new token: %w", err)
	}

	tokens, err := api.RequestJSON[Tokens](logPrefix, client, req)
	if err != nil {
		return fmt.Errorf("making request for refresh tokens: %w", err)
	}

	*t = tokens
	timber.Done(logPrefix, "new access token", t.Access)
	return nil
}

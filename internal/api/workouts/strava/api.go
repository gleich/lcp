package strava

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/secrets"
)

type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	Mutex        sync.Mutex
}

var authTokens tokens

func sendStravaRequest[T any](client *http.Client, path string) (T, error) {
	var zero T

	err := authTokens.RefreshIfExpired(client)
	if err != nil {
		return zero, fmt.Errorf("refreshing tokens: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://www.strava.com/%s", strings.TrimLeft(path, "/")),
		nil,
	)
	if err != nil {
		return zero, fmt.Errorf("creating request: %w", err)
	}
	authTokens.Mutex.Lock()
	req.Header.Set("Authorization", "Bearer "+authTokens.AccessToken)
	authTokens.Mutex.Unlock()

	resp, err := api.RequestJSON[T](client, req, logger())
	if err != nil {
		return zero, fmt.Errorf("making strava api request: %w", err)
	}
	return resp, nil
}

func (t *tokens) RefreshIfExpired(client *http.Client) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	if t.ExpiresAt-60 >= time.Now().Unix() || t.AccessToken == "" {
		return nil
	}

	if t.RefreshToken == "" {
		t.RefreshToken = secrets.ENV.StravaRefreshToken
	}

	start := time.Now()
	params := url.Values{
		"client_id":     {secrets.ENV.StravaClientID},
		"client_secret": {secrets.ENV.StravaClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {t.RefreshToken},
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

	tokens, err := api.RequestJSON[tokens](client, req, logger())
	if err != nil {
		return fmt.Errorf("making request for refresh tokens: %w", err)
	}

	t.AccessToken = tokens.AccessToken
	t.RefreshToken = tokens.RefreshToken
	t.ExpiresAt = tokens.ExpiresAt
	logger().Info().
		Dur("duration", time.Since(start)).
		Str("token", t.AccessToken).
		Msg("new access token")
	return nil
}

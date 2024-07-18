package strava

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gleich/lcp/pkg/secrets"
	"github.com/gleich/lumber/v2"
)

type Tokens struct {
	Access    string `json:"access_token"`
	Refresh   string `json:"refresh_token"`
	ExpiresAt int64  `json:"expires_at"`
}

func LoadTokens(loadedSecrets secrets.Secrets) Tokens {
	return Tokens{
		Access:    loadedSecrets.StravaAccessToken,
		Refresh:   loadedSecrets.StravaRefreshToken,
		ExpiresAt: loadedSecrets.StravaRefreshTokenExpiration,
	}
}

func (t *Tokens) RefreshIfNeeded(loadedSecrets secrets.Secrets) {
	if t.ExpiresAt >= time.Now().Unix() {
		lumber.Debug("Not refreshing token")
		return
	}

	params := url.Values{
		"client_id":     {loadedSecrets.StravaClientID},
		"client_secret": {loadedSecrets.StravaClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {t.Refresh},
		"code":          {loadedSecrets.StravaOAuthCode},
	}
	req, err := http.NewRequest("POST", "https://www.strava.com/oauth/token?"+params.Encode(), nil)
	if err != nil {
		lumber.Error(err, "creating request for new token failed")
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		lumber.Error(err, "sending request for new data token failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		lumber.Error(err, "reading response body failed")
		return
	}
	if resp.StatusCode != http.StatusOK {
		lumber.ErrorMsg(resp.StatusCode, "when trying to get new token data:", string(body))
		return
	}

	var tokens Tokens
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		lumber.Error(err, "failed to parse json")
		lumber.Debug("body:", string(body))
		return
	}

	os.Setenv("STRAVA_ACCESS_TOKEN", tokens.Access)
	os.Setenv("STRAVA_REFRESH_TOKEN", tokens.Refresh)
	os.Setenv("STRAVA_REFRESH_TOKEN_EXPIRATION", strconv.FormatInt(tokens.ExpiresAt, 10))
	t = &tokens

	lumber.Success("loaded new strava token data. access:", t.Access, "refresh:", t.Refresh)
}

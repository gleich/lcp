package secrets

type Secrets struct {
	ValidToken string `env:"VALID_TOKEN"`

	StravaClientID               string `env:"STRAVA_CLIENT_ID"`
	StravaClientSecret           string `env:"STRAVA_CLIENT_SECRET"`
	StravaOAuthCode              string `env:"STRAVA_OAUTH_CODE"`
	StravaAccessToken            string `env:"STRAVA_ACCESS_TOKEN"`
	StravaRefreshToken           string `env:"STRAVA_REFRESH_TOKEN"`
	StravaRefreshTokenExpiration int64  `env:"STRAVA_REFRESH_TOKEN_EXPIRATION"`
}

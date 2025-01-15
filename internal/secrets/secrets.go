package secrets

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"pkg.mattglei.ch/timber"
)

var SECRETS Secrets

type Secrets struct {
	ValidToken  string `env:"VALID_TOKEN"`
	CacheFolder string `env:"CACHE_FOLDER"`

	StravaClientID       string `env:"STRAVA_CLIENT_ID"`
	StravaClientSecret   string `env:"STRAVA_CLIENT_SECRET"`
	StravaOAuthCode      string `env:"STRAVA_OAUTH_CODE"`
	StravaAccessToken    string `env:"STRAVA_ACCESS_TOKEN"`
	StravaRefreshToken   string `env:"STRAVA_REFRESH_TOKEN"`
	StravaSubscriptionID int64  `env:"STRAVA_SUBSCRIPTION_ID"`
	StravaVerifyToken    string `env:"STRAVA_VERIFY_TOKEN"`
	MapboxAccessToken    string `env:"MAPBOX_ACCESS_TOKEN"`
	MinioEndpoint        string `env:"MINIO_ENDPOINT"`
	MinioAccessKeyID     string `env:"MINIO_ACCESS_KEY_ID"`
	MinioSecretKey       string `env:"MINIO_SECRET_KEY"`

	SteamKey string `env:"STEAM_KEY"`
	SteamID  string `env:"STEAM_ID"`

	GitHubAccessToken string `env:"GITHUB_ACCESS_TOKEN"`

	AppleMusicAppToken  string `env:"APPLE_MUSIC_APP_TOKEN"`
	AppleMusicUserToken string `env:"APPLE_MUSIC_USER_TOKEN"`
}

func Load() {
	err := godotenv.Load()
	if err != nil {
		timber.Fatal(err, "loading .env file failed")
	}
	loadedSecrets, err := env.ParseAs[Secrets]()
	if err != nil {
		timber.Fatal(err, "parsing required env vars failed")
	}
	SECRETS = loadedSecrets
	timber.Done("loaded secrets")
}

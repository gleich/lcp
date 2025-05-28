package secrets

import (
	"errors"
	"io/fs"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"go.mattglei.ch/timber"
)

var ENV Secrets

type Secrets struct {
	ValidTokens string `env:"VALID_TOKENS"`
	CacheFolder string `env:"CACHE_FOLDER"`

	// strava
	StravaClientID       string `env:"STRAVA_CLIENT_ID"`
	StravaClientSecret   string `env:"STRAVA_CLIENT_SECRET"`
	StravaOAuthCode      string `env:"STRAVA_OAUTH_CODE"`
	StravaAccessToken    string `env:"STRAVA_ACCESS_TOKEN"`
	StravaRefreshToken   string `env:"STRAVA_REFRESH_TOKEN"`
	StravaSubscriptionID int64  `env:"STRAVA_SUBSCRIPTION_ID"`
	StravaVerifyToken    string `env:"STRAVA_VERIFY_TOKEN"`
	MapboxAccessToken    string `env:"MAPBOX_ACCESS_TOKEN"`

	// hevy
	HevyAccessToken   string  `env:"HEVY_ACCESS_TOKEN"`
	HevyBodyWeightLBS float64 `env:"HEVY_BODY_WEIGHT_LBS"`

	// steam
	SteamKey string `env:"STEAM_KEY"`
	SteamID  string `env:"STEAM_ID"`

	// github
	GitHubAccessToken string `env:"GITHUB_ACCESS_TOKEN"`

	// apple music
	AppleMusicAppToken  string `env:"APPLE_MUSIC_APP_TOKEN"`
	AppleMusicUserToken string `env:"APPLE_MUSIC_USER_TOKEN"`

	// minio
	MinioEndpoint    string `env:"MINIO_ENDPOINT"`
	MinioAccessKeyID string `env:"MINIO_ACCESS_KEY_ID"`
	MinioSecretKey   string `env:"MINIO_SECRET_KEY"`

	// redis
	RedisAddress  string `env:"REDIS_ADDRESS"`
	RedisPassword string `env:"REDIS_PASSWORD"`
}

func Load() {
	if _, err := os.Stat(".env"); !errors.Is(err, fs.ErrNotExist) {
		err := godotenv.Load()
		if err != nil {
			timber.Fatal(err, "loading .env file failed")
		}
	}

	secrets, err := env.ParseAs[Secrets]()
	if err != nil {
		timber.Fatal(err, "parsing required env vars failed")
	}
	ENV = secrets
	timber.Done("loaded secrets")
}

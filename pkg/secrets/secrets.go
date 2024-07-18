package secrets

type Secrets struct {
	ValidToken        string `env:"VALID_TOKEN"`
	StravaAccessToken string `env:"STRAVA_ACCESS_TOKEN"`
}

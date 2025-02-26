package activities

import (
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.mattglei.ch/lcp-2/internal/apis/activities/strava"
	"go.mattglei.ch/lcp-2/internal/cache"
	"go.mattglei.ch/lcp-2/internal/secrets"
	"go.mattglei.ch/timber"
)

const LOG_PREFIX = "[activities]"

func Setup(mux *http.ServeMux) {
	client := http.Client{}
	stravaTokens := strava.LoadTokens()
	err := stravaTokens.RefreshIfNeeded(&client)
	if err != nil {
		timber.Error(err, "failed to refresh strava token data on boot")
	}
	minioClient, err := minio.New(secrets.ENV.MinioEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			secrets.ENV.MinioAccessKeyID,
			secrets.ENV.MinioSecretKey,
			"",
		),
		Secure: true,
	})
	if err != nil {
		timber.Fatal(err, "failed to create minio client")
	}
	activities, err := fetch(&client, *minioClient, stravaTokens)
	if err != nil {
		timber.Error(err, "failed to load initial data for workouts cache; not updating")
	}
	activityCache := cache.New("workouts", activities, err == nil)

	mux.HandleFunc("GET /activities", activityCache.ServeHTTP)
	mux.HandleFunc(
		"POST /strava/event",
		strava.EventRoute(&client, activityCache, *minioClient, fetch, stravaTokens),
	)
	mux.HandleFunc("GET /strava/event", strava.ChallengeRoute)

	timber.Done(LOG_PREFIX, "setup cache and endpoints")
}

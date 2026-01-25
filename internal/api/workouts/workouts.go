package workouts

import (
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/api/workouts/strava"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/timber"
)

const cacheInstance = cache.Workouts

func Setup(mux *http.ServeMux, client *http.Client, rdb *redis.Client) {
	stravaTokens := strava.LoadTokens()
	err := stravaTokens.RefreshIfExpired(client)
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
	activities, err := fetch(client, minioClient, rdb, stravaTokens)
	if err != nil {
		timber.Error(err, "failed to load initial data for workouts cache; not updating")
	}
	workoutsCache := cache.New(cacheInstance, activities, err == nil)

	workoutsCache.Endpoints(mux)
	mux.HandleFunc(
		"POST /strava/event",
		strava.EventRoute(client, workoutsCache, minioClient, rdb, fetch, stravaTokens),
	)
	mux.HandleFunc("GET /strava/event", strava.ChallengeRoute)

	timber.Done(cacheInstance.LogPrefix(), "setup cache and endpoints")
}

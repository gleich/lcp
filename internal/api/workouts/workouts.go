package workouts

import (
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/api/workouts/strava"
	"go.mattglei.ch/lcp/internal/cache"
)

const cacheInstance = cache.Workouts

var logger = cacheInstance.LazyLogger()

func Setup(mux *http.ServeMux, client *http.Client, minioClient *minio.Client, rdb *redis.Client) {
	stravaTokens := strava.LoadTokens()
	err := stravaTokens.RefreshIfExpired(client)
	if err != nil {
		logger().Error().Err(err).Msg("failed to refresh strava token data on boot")
	}
	activities, err := fetch(client, minioClient, rdb, stravaTokens)
	if err != nil {
		logger().Error().
			Err(err).
			Msg("failed to load initial data for workouts cache; not updating")
	}
	workoutsCache := cache.New(cacheInstance, activities, err == nil)

	workoutsCache.Endpoints(mux)
	mux.HandleFunc(
		"POST /strava/event",
		strava.EventRoute(client, workoutsCache, minioClient, rdb, fetch, stravaTokens),
	)
	mux.HandleFunc("GET /strava/event", strava.ChallengeRoute)
}

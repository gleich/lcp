package steam

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/pkg/lcp"
)

const cacheInstance = cache.Steam

var logger = cacheInstance.LazyLogger()

func Setup(mux *http.ServeMux, client *http.Client, rdb *redis.Client) {
	games, err := fetchRecentlyPlayedGames(client, rdb)
	if err != nil {
		logger().Error().Err(err).Msg("initial fetch of steam games failed")
	}

	steamCache := cache.New(cacheInstance, games, err == nil)
	steamCache.Endpoints(mux)
	go cache.UpdatePeriodically(
		steamCache,
		client,
		func(client *http.Client) ([]lcp.SteamGame, error) {
			return fetchRecentlyPlayedGames(client, rdb)
		},
		10*time.Minute,
	)
}

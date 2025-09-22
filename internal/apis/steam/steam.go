package steam

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

const cacheInstance = cache.Steam

func Setup(mux *http.ServeMux, client *http.Client, rdb *redis.Client) {
	games, err := fetchRecentlyPlayedGames(client, rdb)
	if err != nil {
		timber.Error(err, "initial fetch of steam games failed")
	}

	steamCache := cache.New(cacheInstance, games, err == nil)
	mux.HandleFunc("GET /steam", steamCache.Serve)
	go cache.UpdatePeriodically(
		steamCache,
		client,
		func(client *http.Client) ([]lcp.SteamGame, error) {
			return fetchRecentlyPlayedGames(client, rdb)
		},
		15*time.Minute,
	)
	timber.Done(cacheInstance.LogPrefix(), "setup cache and endpoint")
}

package steam

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/tasks"
	"go.mattglei.ch/lcp/pkg/lcp"
)

const cacheInstance = cache.Steam

func Setup(mux *http.ServeMux, client *http.Client, rdb *redis.Client) {
	task, start := tasks.Cache.Steam.Setup.Start()
	games, err := fetchRecentlyPlayedGames(client, rdb)
	if err != nil {
		task.Error(err, "initial fetch of steam games failed")
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
	task.InfoSince("setup cache and endpoint", start)
}

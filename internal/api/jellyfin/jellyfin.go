package jellyfin

import (
	"net/http"
	"time"

	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

const cacheInstance = cache.Jellyfin

var logAttr = cacheInstance.LogAttr()

func Setup(mux *http.ServeMux, client *http.Client) {
	items, err := fetchRecentlyPlayed(client)
	if err != nil {
		timber.Error(err, "initial fetch of steam games failed", logAttr)
	}

	jellyfinCache := cache.New(cacheInstance, items, err == nil)
	jellyfinCache.Endpoints(mux)
	go cache.UpdatePeriodically(
		jellyfinCache,
		client,
		func(client *http.Client) ([]lcp.JellyfinItem, error) {
			return fetchRecentlyPlayed(client)
		},
		1*time.Minute,
	)
}

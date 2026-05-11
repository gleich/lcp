package jellyfin

import (
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

const cacheInstance = cache.Jellyfin

var logAttr = cacheInstance.LogAttr()

func Setup(mux *http.ServeMux, client *http.Client, minioClient *minio.Client) {
	items, err := fetchRecentlyPlayed(client, minioClient)
	if err != nil {
		timber.Error(err, "initial fetch of jellyfin items failed", logAttr)
	}

	jellyfinCache := cache.New(cacheInstance, items, err == nil)
	jellyfinCache.Endpoints(mux)
	go cache.UpdatePeriodically(
		jellyfinCache,
		client,
		func(client *http.Client) ([]lcp.JellyfinItem, error) {
			return fetchRecentlyPlayed(client, minioClient)
		},
		1*time.Minute,
	)
}

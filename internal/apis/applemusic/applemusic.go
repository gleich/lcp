package applemusic

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

const cacheInstance = cache.AppleMusic

func cacheUpdate(client *http.Client, rdb *redis.Client) (lcp.AppleMusicCache, error) {
	recentlyPlayed, err := fetchRecentlyPlayed(client, rdb)
	if err != nil {
		return lcp.AppleMusicCache{}, err
	}

	playlistsIDs := []string{
		"p.V7VYVB0hZo53MQv", // old man
		"p.AWXoZoxHLrvpJlY", // chill
		"p.AWXoXPYSLrvpJlY", // alt
		"p.qQXLX2rHA75zg8e", // after hours
		"p.gek1E8efLa68Adp", // classics
		"p.qQXLxPLtA75zg8e", // 80s
		"p.LV0PXNoCl0EpDLW", // divorced dad
		"p.QvDQE5RIVbAeokL", // party
		"p.LV0PXL3Cl0EpDLW", // bops
		"p.6xZaArOsvzb5OML", // focus
		"p.AWXoXeAiLrvpJlY", // smooth
		"p.O1kz7EoFVmvz704", // funk
		"p.qQXLxPpFA75zg8e", // rap
		"p.qQXLxpDuA75zg8e", // ROCK
		"p.O1kz7zbsVmvz704", // country
		"p.QvDQEN0IVbAeokL", // fall
	}
	playlists := []lcp.AppleMusicPlaylist{}
	for _, id := range playlistsIDs {
		playlistData, err := fetchPlaylist(client, rdb, id)
		if err != nil {
			return lcp.AppleMusicCache{}, err
		}
		playlists = append(playlists, playlistData)
	}

	return lcp.AppleMusicCache{
		RecentlyPlayed: recentlyPlayed,
		Playlists:      playlists,
	}, nil
}

func Setup(mux *http.ServeMux, client *http.Client, rdb *redis.Client) {
	data, err := cacheUpdate(client, rdb)
	if err != nil {
		timber.Error(err, "initial fetch of applemusic cache data failed")
	}

	applemusicCache := cache.New(cacheInstance, data, err == nil)
	mux.HandleFunc("GET /applemusic", applemusicCache.ServeHTTP)
	go cache.UpdatePeriodically(
		applemusicCache,
		client,
		func(client *http.Client) (lcp.AppleMusicCache, error) {
			return cacheUpdate(client, rdb)
		},
		30*time.Second,
	)
	timber.Done(cacheInstance.LogPrefix(), "setup cache and endpoints")
}

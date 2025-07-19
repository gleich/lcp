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

	playlists := map[string]string{
		"p.AWXoZoxHLrvpJlY": "https://open.spotify.com/playlist/5SnoWhWIJRmJNkvdxCpMAe?si=A4x-F7uxRpeTSRWN4kwdRw", // chill
		"p.AWXoXPYSLrvpJlY": "https://open.spotify.com/playlist/7tN57nLbiiw4bliUyw2oYL?si=CE8VaLIgTRu26YS4IvwN-A", // alt
		"p.qQXLX2rHA75zg8e": "https://open.spotify.com/playlist/1NMII2bpE3l7CvBxYVK7Fu?si=MYRsVTZpQz6x1TouvOXb1g", // after hours
		"p.gek1E8efLa68Adp": "https://open.spotify.com/playlist/2HYOAlwB570McLyD3nIJKG?si=QJHGnwRpQgWu6OrFxacq-w", // classics
		"p.qQXLxPLtA75zg8e": "https://open.spotify.com/playlist/1DB0cG12kphRKvNzKPGmpf?si=P9ALlLDzSE2MU2eIrPF0mg", // 80s
		"p.V7VYVB0hZo53MQv": "https://open.spotify.com/playlist/3fDlIqV43BvPvtPs9ASsgU?si=e3QywoGZT76f6MPINPh1JA", // old man
		"p.LV0PXNoCl0EpDLW": "https://open.spotify.com/playlist/3p0bSspMsoZ0QodpDCcb3U?si=Hw-dupHLSGCgqduvB8OWUQ", // divorced dad
		"p.QvDQE5RIVbAeokL": "https://open.spotify.com/playlist/6AFH5WO2uZeSwKdirNvryH?si=lzVfDDkUTHS48f-_5bEDUQ", // party
		"p.LV0PXL3Cl0EpDLW": "https://open.spotify.com/playlist/2Bc0msBHeRaNYUFO8LfHct?si=GrzK7i4BQjmC0hXWScpiCg", // bops
		"p.6xZaArOsvzb5OML": "https://open.spotify.com/playlist/261XCji6XWsXktcMumPIqa?si=WE1l-BPjRG-Li2Sx8znm7Q", // focus
		"p.AWXoXeAiLrvpJlY": "https://open.spotify.com/playlist/2CvjwmuE5CekSZ1CfezOQO?si=C8m53iq-RW20_KRylT5-sw", // smooth
		"p.O1kz7EoFVmvz704": "https://open.spotify.com/playlist/1EDwymox6cXQlk7JGDMCbz?si=f-I1H6duSxuXtLy9K_gBqA", // funk
		"p.qQXLxPpFA75zg8e": "https://open.spotify.com/playlist/6MLAGkQPdSBMjit5O1hrws?si=V5n6ge43SruIycU6mRsQSQ", // rap
		"p.qQXLxpDuA75zg8e": "https://open.spotify.com/playlist/1Yh42cCAPWPFy55hw8VaWJ?si=7uj8ojHxQhWrakmKn6DyOg", // rock
		"p.O1kz7zbsVmvz704": "https://open.spotify.com/playlist/3jR0MH0NwzEdYuUY8nohmf?si=u5pJqEM5TUu3t8IVo5XNxQ", // country
		"p.QvDQEN0IVbAeokL": "https://open.spotify.com/playlist/2QZIyEGgLToF97wqMgU9yZ?si=N-mX-sXNS4W0_SIEJYAzpw", // fall
	}
	appleMusicPlaylists := []lcp.AppleMusicPlaylist{}
	for id, spotifyURL := range playlists {
		playlistData, err := fetchPlaylist(client, rdb, id, spotifyURL)
		if err != nil {
			return lcp.AppleMusicCache{}, err
		}
		appleMusicPlaylists = append(appleMusicPlaylists, playlistData)
	}

	return lcp.AppleMusicCache{
		RecentlyPlayed: recentlyPlayed,
		Playlists:      appleMusicPlaylists,
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

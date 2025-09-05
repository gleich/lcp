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

type playlist struct {
	ID         string
	SpotifyURL string
}

func cacheUpdate(client *http.Client, rdb *redis.Client) (lcp.AppleMusicCache, error) {
	recentlyPlayed, err := fetchRecentlyPlayed(client, rdb)
	if err != nil {
		return lcp.AppleMusicCache{}, err
	}

	playlists := []playlist{
		// chill
		{
			ID:         "p.AWXoZoxHLrvpJlY",
			SpotifyURL: "https://open.spotify.com/playlist/5SnoWhWIJRmJNkvdxCpMAe?si=A4x-F7uxRpeTSRWN4kwdRw",
		},
		// alt
		{
			ID:         "p.AWXoXPYSLrvpJlY",
			SpotifyURL: "https://open.spotify.com/playlist/7tN57nLbiiw4bliUyw2oYL?si=CE8VaLIgTRu26YS4IvwN-A",
		},
		// after hours
		{
			ID:         "p.qQXLX2rHA75zg8e",
			SpotifyURL: "https://open.spotify.com/playlist/1NMII2bpE3l7CvBxYVK7Fu?si=MYRsVTZpQz6x1TouvOXb1g",
		},
		// smooth
		{
			ID:         "p.AWXoXeAiLrvpJlY",
			SpotifyURL: "https://open.spotify.com/playlist/2CvjwmuE5CekSZ1CfezOQO?si=C8m53iq-RW20_KRylT5-sw",
		},
		// classics
		{
			ID:         "p.gek1E8efLa68Adp",
			SpotifyURL: "https://open.spotify.com/playlist/2HYOAlwB570McLyD3nIJKG?si=QJHGnwRpQgWu6OrFxacq-w",
		},
		// 80s
		{
			ID:         "p.qQXLxPLtA75zg8e",
			SpotifyURL: "https://open.spotify.com/playlist/1DB0cG12kphRKvNzKPGmpf?si=P9ALlLDzSE2MU2eIrPF0mg",
		},
		// divorced dad
		{
			ID:         "p.LV0PXNoCl0EpDLW",
			SpotifyURL: "https://open.spotify.com/playlist/3p0bSspMsoZ0QodpDCcb3U?si=Hw-dupHLSGCgqduvB8OWUQ",
		},
		// party
		{
			ID:         "p.QvDQE5RIVbAeokL",
			SpotifyURL: "https://open.spotify.com/playlist/6AFH5WO2uZeSwKdirNvryH?si=lzVfDDkUTHS48f-_5bEDUQ",
		},
		// bops
		{
			ID:         "p.LV0PXL3Cl0EpDLW",
			SpotifyURL: "https://open.spotify.com/playlist/2Bc0msBHeRaNYUFO8LfHct?si=GrzK7i4BQjmC0hXWScpiCg",
		},
		// house
		{
			ID:         "p.gek1EWzCLa68Adp",
			SpotifyURL: "https://open.spotify.com/playlist/3iMi8ew4XvYCCcS9P2iARw",
		},
		// focus
		{
			ID:         "p.6xZaArOsvzb5OML",
			SpotifyURL: "https://open.spotify.com/playlist/261XCji6XWsXktcMumPIqa?si=WE1l-BPjRG-Li2Sx8znm7Q",
		},
		// funk
		{
			ID:         "p.O1kz7EoFVmvz704",
			SpotifyURL: "https://open.spotify.com/playlist/1EDwymox6cXQlk7JGDMCbz?si=f-I1H6duSxuXtLy9K_gBqA",
		},
		// old man
		{
			ID:         "p.V7VYVB0hZo53MQv",
			SpotifyURL: "https://open.spotify.com/playlist/3fDlIqV43BvPvtPs9ASsgU?si=e3QywoGZT76f6MPINPh1JA",
		},
		// rap
		{
			ID:         "p.qQXLxPpFA75zg8e",
			SpotifyURL: "https://open.spotify.com/playlist/6MLAGkQPdSBMjit5O1hrws?si=V5n6ge43SruIycU6mRsQSQ",
		},
		// rock
		{
			ID:         "p.qQXLxpDuA75zg8e",
			SpotifyURL: "https://open.spotify.com/playlist/1Yh42cCAPWPFy55hw8VaWJ?si=7uj8ojHxQhWrakmKn6DyOg",
		},
		// country
		{
			ID:         "p.O1kz7zbsVmvz704",
			SpotifyURL: "https://open.spotify.com/playlist/3jR0MH0NwzEdYuUY8nohmf?si=u5pJqEM5TUu3t8IVo5XNxQ",
		},
	}
	appleMusicPlaylists := []lcp.AppleMusicPlaylist{}
	for _, playlist := range playlists {
		playlistData, err := fetchPlaylist(client, rdb, playlist.ID, playlist.SpotifyURL)
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

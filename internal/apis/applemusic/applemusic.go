package applemusic

import (
	"encoding/json"
	"fmt"
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

	appleMusicPlaylists := []lcp.AppleMusicPlaylist{}
	for _, playlist := range playlists {
		playlistData, err := fetchPlaylist(client, rdb, playlist)
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
	applemusicCache.MarshalResponse = MarshalResponse
	applemusicCache.Endpoints(mux)
	mux.HandleFunc("GET /applemusic/playlists", syncedPlaylistsEndpoint())
	mux.HandleFunc("GET /applemusic/playlists/{id}", playlistEndpoint(applemusicCache))
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

func MarshalResponse(
	c *cache.Cache[lcp.AppleMusicCache],
) (string, error) {
	response := lcp.CacheResponse[lcp.AppleMusicCacheResponse]{Updated: c.Updated}
	response.Data.RecentlyPlayed = c.Data.RecentlyPlayed
	for _, p := range c.Data.Playlists {
		firstFourTracks := []lcp.AppleMusicSong{}
		for _, track := range p.Tracks {
			if len(firstFourTracks) < 4 {
				firstFourTracks = append(firstFourTracks, track)
			}
		}
		response.Data.PlaylistSummaries = append(
			response.Data.PlaylistSummaries,
			lcp.AppleMusicPlaylistSummary{
				Name:            p.Name,
				ID:              p.ID,
				TrackCount:      len(p.Tracks),
				FirstFourTracks: firstFourTracks,
			},
		)
	}

	data, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("%w failed to json encode data", err)
	}
	return string(data), nil
}

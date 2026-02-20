package applemusic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/util"
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
	start := time.Now()
	data, err := cacheUpdate(client, rdb)
	if err != nil {
		timber.Error(err, "initial fetch of applemusic cache data failed")
	}

	applemusicCache := cache.New(cacheInstance, data, err == nil)
	applemusicCache.MarshalResponse = marshalResponse
	applemusicCache.DiffCheck = diffCheck
	applemusicCache.Endpoints(mux)
	mux.HandleFunc("GET /applemusic/playlists/{id}", playlistEndpoint(applemusicCache))
	go cache.UpdatePeriodically(
		applemusicCache,
		client,
		func(client *http.Client) (lcp.AppleMusicCache, error) {
			return cacheUpdate(client, rdb)
		},
		10*time.Second,
	)
	timber.DoneSince(start, cacheInstance.LogPrefix(), "setup cache and endpoints")
}

func marshalResponse(
	c *cache.Cache[lcp.AppleMusicCache],
) ([]byte, error) {
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
		return []byte{}, fmt.Errorf("encoding json data: %w", err)
	}
	return data, nil
}

// We need a custom diff check function due to the fact that apple music image service returns
// images with the permissions attached as url params. These change every time we make a request for
// the playlist. This was causing false updates, so this will check for differences in everything,
// normalize the album art URL for the url without the permissions, and check to see if the
// permissions are expired.
func diffCheck(c *cache.Cache[lcp.AppleMusicCache], old, new lcp.AppleMusicCache) (bool, error) {
	different, err := diffSongList(old.RecentlyPlayed, new.RecentlyPlayed)
	if err != nil {
		return false, fmt.Errorf("recently played diff check: %w", err)
	}
	if different {
		return true, nil
	}

	if len(old.Playlists) != len(new.Playlists) {
		return true, nil
	}
	for i, oldPlaylist := range old.Playlists {
		newPlaylist := new.Playlists[i]
		different, err = diffSongList(oldPlaylist.Tracks, newPlaylist.Tracks)
		if err != nil {
			return false, fmt.Errorf("checking track difference for %s: %w", oldPlaylist.Name, err)
		}
		if different {
			return true, nil
		}
		if oldPlaylist.Name != newPlaylist.Name ||
			oldPlaylist.LastModified != newPlaylist.LastModified ||
			oldPlaylist.URL != newPlaylist.URL ||
			oldPlaylist.SpotifyID != newPlaylist.SpotifyID ||
			oldPlaylist.ID != newPlaylist.ID {
			return true, nil
		}
	}

	return false, nil
}

func diffSongList(oldSongs, newSongs []lcp.AppleMusicSong) (bool, error) {
	if len(oldSongs) != len(newSongs) {
		return true, nil
	}

	for i, old := range oldSongs {
		new := newSongs[i]

		if old.Track != new.Track || old.Artist != new.Artist ||
			old.DurationInMillis != new.DurationInMillis ||
			old.URL != new.URL ||
			old.PreviewAudioURL != new.PreviewAudioURL {
			return true, nil
		}

		if old.AlbumArtURL != nil && new.AlbumArtURL != nil {
			if old.AlbumArtPermissionsExpiration != nil &&
				time.Now().After(old.AlbumArtPermissionsExpiration.Add(-1*time.Minute)) {
				var (
					oldURL = old.AlbumArtURL
					newURL = new.AlbumArtURL
				)
				if old.AlbumArtURL == newURL {
					timber.Warning("album art did not update even though it expired")
					timber.Warning("old url:", oldURL)
					timber.Warning("new url:", newURL)
				} else {
					return true, nil
				}
			}

			oldArtURL, err := util.NormalizeURL(*old.AlbumArtURL)
			if err != nil {
				return false, fmt.Errorf(
					"failed to normalize url for %s: %w",
					*old.AlbumArtURL,
					err,
				)
			}
			newArtURL, err := util.NormalizeURL(*new.AlbumArtURL)
			if err != nil {
				return false, fmt.Errorf(
					"failed to normalize url for %s: %w",
					*new.AlbumArtURL,
					err,
				)
			}

			if oldArtURL.String() != newArtURL.String() {
				return true, nil
			}
		}
	}

	return false, nil
}

package applemusic

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/pkg/lcp"
)

type recentlyPlayedResponse struct {
	Data []songResponse `json:"data"`
}

const recentlyPlayedLimit = 10

func fetchRecentlyPlayed(
	client *http.Client,
	rdb *redis.Client,
	blacklist *blacklistCache,
) ([]lcp.AppleMusicSong, error) {
	params := url.Values{
		"types": {"songs,library-songs"},
	}
	response, err := sendAppleMusicRequest[recentlyPlayedResponse](
		client,
		"/v1/me/recent/played/tracks"+params.Encode(),
	)
	if err != nil {
		return []lcp.AppleMusicSong{}, err
	}

	var songs []lcp.AppleMusicSong
	for _, s := range response.Data {
		so, err := s.ToAppleMusicSong(client, rdb)
		if err != nil {
			return []lcp.AppleMusicSong{}, fmt.Errorf(
				"parsing song from song response: %w",
				err,
			)
		}
		songs = append(songs, so)
	}

	// filter out duplicate songs
	seen := make(map[string]bool)
	uniqueSongs := []lcp.AppleMusicSong{}
	for _, song := range songs {
		if !seen[song.ID] {
			seen[song.ID] = true
			uniqueSongs = append(uniqueSongs, song)
		}
	}

	// remove and replace songs from blacklist
	blacklist.Mutex.RLock()
	defer blacklist.Mutex.RUnlock()

	filteredSongs := []lcp.AppleMusicSong{}
	poolIndex := 0
	for _, song := range uniqueSongs[:min(len(uniqueSongs), recentlyPlayedLimit)] {
		if !blacklist.Contains(song) {
			filteredSongs = append(filteredSongs, song)
			continue
		}

		for poolIndex < len(blacklist.ReplacementPool) &&
			seen[blacklist.ReplacementPool[poolIndex].ID] {
			poolIndex++
		}
		if poolIndex >= len(blacklist.ReplacementPool) {
			logger().Warn().
				Str("song", song.ID).
				Msg("no replacement left in pool for blacklisted song")
			continue
		}

		replacement := blacklist.ReplacementPool[poolIndex]
		seen[replacement.ID] = true
		poolIndex++
		filteredSongs = append(filteredSongs, replacement)
	}

	return filteredSongs, nil
}

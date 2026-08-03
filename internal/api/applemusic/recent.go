package applemusic

import (
	"fmt"
	"net/http"

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
	response, err := sendAppleMusicRequest[recentlyPlayedResponse](
		client,
		"/v1/me/recent/played/tracks",
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
	blacklistedIndex := 0
	for _, song := range uniqueSongs[:recentlyPlayedLimit] {
		if !blacklist.Contains(song) {
			filteredSongs = append(filteredSongs, song)
			continue
		}

		for blacklistedIndex < len(blacklist.ReplacementPool) &&
			seen[blacklist.ReplacementPool[blacklistedIndex].ID] {
			blacklistedIndex++
		}
		if blacklistedIndex >= len(blacklist.ReplacementPool) {
			return []lcp.AppleMusicSong{}, fmt.Errorf(
				"no replacement left in pool for blacklisted song %s",
				song.ID,
			)
		}

		replacement := blacklist.ReplacementPool[blacklistedIndex]
		seen[replacement.ID] = true
		blacklistedIndex++
		filteredSongs = append(filteredSongs, replacement)
	}

	return filteredSongs, nil
}

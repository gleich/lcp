package applemusic

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/pkg/lcp"
)

type blacklistCache struct {
	Mutex           sync.RWMutex
	Songs           []lcp.AppleMusicSong
	ReplacementPool []lcp.AppleMusicSong
}

// Contains reports if a song is blacklisted. Library and catalog versions of
// the same track can carry different IDs, so fall back to track and artist.
// Callers must hold the read lock.
func (b *blacklistCache) Contains(song lcp.AppleMusicSong) bool {
	return slices.ContainsFunc(b.Songs, func(s lcp.AppleMusicSong) bool {
		return s.ID == song.ID || songIdentity(s) == songIdentity(song)
	})
}

func songIdentity(s lcp.AppleMusicSong) string {
	return strings.ToLower(s.Track) + "|" + strings.ToLower(s.Artist)
}

func (b *blacklistCache) Refresh(client *http.Client, rdb *redis.Client) error {
	const replacementPool = "p.PkxVvRoH2zv9xE8" // sleep sounds
	pool, err := fetchPlaylistTracks(client, rdb, replacementPool)
	if err != nil {
		return fmt.Errorf("fetching replacement pool: %w", err)
	}
	if len(pool) == 0 {
		logger().Warn().Msg("replacement pool is empty; blacklisted songs will be dropped")
	}

	playlists := []string{
		"p.PkxVxXei2zv9xE8", // ttmwf
		"p.1YeWLY8hqkred4L", // test
	}
	var songs []lcp.AppleMusicSong
	for _, id := range playlists {
		playlistSongs, err := fetchPlaylistTracks(client, rdb, id)
		if err != nil {
			return fmt.Errorf("fetching blacklist playlist %s: %w", id, err)
		}
		songs = append(songs, playlistSongs...)
	}

	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	b.ReplacementPool = pool
	b.Songs = songs
	return nil
}

func (b *blacklistCache) UpdatePeriodically(client *http.Client, rdb *redis.Client) {
	for {
		time.Sleep(5 * time.Minute)
		err := b.Refresh(client, rdb)
		if err != nil {
			logger().Error().Err(err).Msg("updating blacklist failed")
		}
	}
}

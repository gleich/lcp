package applemusic

import (
	"net/http"
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

func (b *blacklistCache) Refresh(client *http.Client, rdb *redis.Client) error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	// refreshing pool
	const source = "p.PkxVvRoH2zv9xE8" // sleep sounds
	poolSongs, err := fetchPlaylistTracks(client, rdb, source)
	if err != nil {
		return err
	}
	b.ReplacementPool = poolSongs

	// refreshing songs
	playlists := []string{
		"p.PkxVxXei2zv9xE8", // ttmwf
		"p.1YeWLY8hqkred4L", // test
	}
	var songs []lcp.AppleMusicSong
	for _, id := range playlists {
		playlistSongs, err := fetchPlaylistTracks(client, rdb, id)
		if err != nil {
			return err
		}
		songs = append(songs, playlistSongs...)
	}
	b.Songs = songs
	return nil
}

func (b *blacklistCache) UpdatePeriodically(client *http.Client, rdb *redis.Client) error {
	for {
		time.Sleep(5 * time.Minute)
		err := b.Refresh(client, rdb)
		if err != nil {
			logger().Error().Err(err).Msg("updating blacklist failed")
		}
	}
}

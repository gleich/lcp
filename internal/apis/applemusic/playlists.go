package applemusic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/auth"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

var (
	playlists = []lcp.AppleMusicSyncedPlaylist{
		{Name: "chill", AppleMusicID: "p.AWXoZoxHLrvpJlY", SpotifyID: "5SnoWhWIJRmJNkvdxCpMAe"},
		{
			Name:         "after hours",
			AppleMusicID: "p.qQXLX2rHA75zg8e",
			SpotifyID:    "1NMII2bpE3l7CvBxYVK7Fu",
		},
		{Name: "bops", AppleMusicID: "p.LV0PXL3Cl0EpDLW", SpotifyID: "2Bc0msBHeRaNYUFO8LfHct"},
		{Name: "smooth", AppleMusicID: "p.AWXoXeAiLrvpJlY", SpotifyID: "2CvjwmuE5CekSZ1CfezOQO"},
		{Name: "classics", AppleMusicID: "p.gek1E8efLa68Adp", SpotifyID: "2HYOAlwB570McLyD3nIJKG"},
		{Name: "80s", AppleMusicID: "p.qQXLxPLtA75zg8e", SpotifyID: "1DB0cG12kphRKvNzKPGmpf"},
		{Name: "alt", AppleMusicID: "p.AWXoXPYSLrvpJlY", SpotifyID: "7tN57nLbiiw4bliUyw2oYL"},
		{
			Name:         "divorced dad",
			AppleMusicID: "p.LV0PXNoCl0EpDLW",
			SpotifyID:    "3p0bSspMsoZ0QodpDCcb3U",
		},
		{Name: "party", AppleMusicID: "p.QvDQE5RIVbAeokL", SpotifyID: "6AFH5WO2uZeSwKdirNvryH"},
		{Name: "house", AppleMusicID: "p.gek1EWzCLa68Adp", SpotifyID: "3iMi8ew4XvYCCcS9P2iARw"},
		{Name: "focus", AppleMusicID: "p.6xZaArOsvzb5OML", SpotifyID: "261XCji6XWsXktcMumPIqa"},
		{Name: "funk", AppleMusicID: "p.O1kz7EoFVmvz704", SpotifyID: "1EDwymox6cXQlk7JGDMCbz"},
		{Name: "old man", AppleMusicID: "p.V7VYVB0hZo53MQv", SpotifyID: "3fDlIqV43BvPvtPs9ASsgU"},
		{Name: "quiet", AppleMusicID: "p.AWXoXbWfLrvpJlY", SpotifyID: "6ex8usE1U41m4wM6NwT9cg"},
		{Name: "rock", AppleMusicID: "p.qQXLxpDuA75zg8e", SpotifyID: "1Yh42cCAPWPFy55hw8VaWJ"},
		{Name: "country", AppleMusicID: "p.O1kz7zbsVmvz704", SpotifyID: "3jR0MH0NwzEdYuUY8nohmf"},
	}
)

type playlistTracksResponse struct {
	Next string         `json:"next"`
	Data []songResponse `json:"data"`
}

type playlistResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			LastModifiedDate time.Time `json:"lastModifiedDate"`
			Name             string    `json:"name"`
			PlayParams       struct {
				GlobalID string `json:"globalId"`
			} `json:"playParams"`
		} `json:"attributes"`
	} `json:"data"`
}

func fetchPlaylist(
	client *http.Client,
	rdb *redis.Client,
	playlist lcp.AppleMusicSyncedPlaylist,
) (lcp.AppleMusicPlaylist, error) {
	playlistData, err := sendAppleMusicAPIRequest[playlistResponse](
		client,
		fmt.Sprintf("/v1/me/library/playlists/%s", playlist.AppleMusicID),
	)
	if err != nil {
		return lcp.AppleMusicPlaylist{}, fmt.Errorf(
			"%w failed to fetch playlist for %s",
			err,
			playlist.AppleMusicID,
		)
	}

	var tracks []lcp.AppleMusicSong
	path := fmt.Sprintf("/v1/me/library/playlists/%s/tracks", playlist.AppleMusicID)
	for {
		trackData, err := sendAppleMusicAPIRequest[playlistTracksResponse](client, path)
		if err != nil {
			return lcp.AppleMusicPlaylist{}, fmt.Errorf(
				"%w failed to fetch playlist data for %s",
				err,
				path,
			)
		}
		for _, track := range trackData.Data {
			song, err := songFromSongResponse(client, rdb, track)
			if err != nil {
				return lcp.AppleMusicPlaylist{}, fmt.Errorf(
					"%w failed create song from apple music song response",
					err,
				)
			}
			tracks = append(tracks, song)
		}

		if trackData.Next == "" {
			break
		}
		path = trackData.Next
	}

	return lcp.AppleMusicPlaylist{
		Name:         playlistData.Data[0].Attributes.Name,
		LastModified: playlistData.Data[0].Attributes.LastModifiedDate,
		Tracks:       tracks,
		ID:           playlistData.Data[0].ID,
		URL: fmt.Sprintf(
			"https://music.apple.com/us/playlist/alt/%s",
			playlistData.Data[0].Attributes.PlayParams.GlobalID,
		),
		SpotifyID: playlist.SpotifyID,
	}, nil
}

func syncedPlaylistsEndpoint() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAuthorized(w, r) {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(playlists)
		if err != nil {
			err = fmt.Errorf("%w failed to write json data to request", err)
			timber.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func playlistEndpoint(c *cache.Cache[lcp.AppleMusicCache]) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAuthorized(w, r) {
			return
		}

		id := r.PathValue("id")

		c.Mutex.RLock()
		var p *lcp.AppleMusicPlaylist
		for _, plist := range c.Data.Playlists {
			if plist.ID == id {
				p = &plist
				break
			}
		}

		if p == nil {
			c.Mutex.RUnlock()
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(p)
		c.Mutex.RUnlock()
		if err != nil {
			err = fmt.Errorf("%w failed to write json data to request", err)
			timber.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

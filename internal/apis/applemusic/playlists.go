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
	playlists = []lcp.SyncedPlaylist{
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
		// quiet
		{
			ID:         "p.AWXoXbWfLrvpJlY",
			SpotifyURL: "",
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
	id string,
	spotifyURL string,
) (lcp.AppleMusicPlaylist, error) {
	playlistData, err := sendAppleMusicAPIRequest[playlistResponse](
		client,
		fmt.Sprintf("/v1/me/library/playlists/%s", id),
	)
	if err != nil {
		return lcp.AppleMusicPlaylist{}, fmt.Errorf(
			"%w failed to fetch playlist for %s",
			err,
			id,
		)
	}

	var tracks []lcp.AppleMusicSong
	path := fmt.Sprintf("/v1/me/library/playlists/%s/tracks", id)
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
		SpotifyURL: spotifyURL,
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

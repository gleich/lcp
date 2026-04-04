package applemusic

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/auth"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/util"
	"go.mattglei.ch/lcp/pkg/lcp"
)

type syncedPlaylist struct {
	Name         string
	AppleMusicID string
	SpotifyID    string
}

var playlists = []syncedPlaylist{
	// {Name: "christmas", AppleMusicID: "p.QvDQEebsVbAeokL", SpotifyID: "4sxPVSb9VcA4RQOY7lKQxI"},
	// {
	// 	Name:         "friendsgiving",
	// 	AppleMusicID: "p.gek1krvFLa68Adp",
	// 	SpotifyID:    "7IbaiRMhet4tMO0zm7wcds",
	// },
	{Name: "chill", AppleMusicID: "p.AWXoZoxHLrvpJlY", SpotifyID: "5SnoWhWIJRmJNkvdxCpMAe"},
	{Name: "smooth", AppleMusicID: "p.AWXoXeAiLrvpJlY", SpotifyID: "2CvjwmuE5CekSZ1CfezOQO"},
	{Name: "bops", AppleMusicID: "p.LV0PXL3Cl0EpDLW", SpotifyID: "2Bc0msBHeRaNYUFO8LfHct"},
	{
		Name:         "after hours",
		AppleMusicID: "p.qQXLX2rHA75zg8e",
		SpotifyID:    "1NMII2bpE3l7CvBxYVK7Fu",
	},
	{Name: "classics", AppleMusicID: "p.gek1E8efLa68Adp", SpotifyID: "2HYOAlwB570McLyD3nIJKG"},
	{Name: "wii resort", AppleMusicID: "p.ZOAXx8zt4KMD6ob", SpotifyID: "7uvn0NulDH3me9WoPZY2nD"},
	{Name: "80s", AppleMusicID: "p.qQXLxPLtA75zg8e", SpotifyID: "1DB0cG12kphRKvNzKPGmpf"},
	{Name: "alt", AppleMusicID: "p.AWXoXPYSLrvpJlY", SpotifyID: "7tN57nLbiiw4bliUyw2oYL"},
	{
		Name:         "divorced dad",
		AppleMusicID: "p.LV0PXNoCl0EpDLW",
		SpotifyID:    "3p0bSspMsoZ0QodpDCcb3U",
	},
	{Name: "party", AppleMusicID: "p.QvDQE5RIVbAeokL", SpotifyID: "6AFH5WO2uZeSwKdirNvryH"},
	{Name: "house", AppleMusicID: "p.gek1EWzCLa68Adp", SpotifyID: "3iMi8ew4XvYCCcS9P2iARw"},
	{Name: "funk", AppleMusicID: "p.O1kz7EoFVmvz704", SpotifyID: "1EDwymox6cXQlk7JGDMCbz"},
	{Name: "old man", AppleMusicID: "p.V7VYVB0hZo53MQv", SpotifyID: "3fDlIqV43BvPvtPs9ASsgU"},
	{Name: "country", AppleMusicID: "p.O1kz7zbsVmvz704", SpotifyID: "3jR0MH0NwzEdYuUY8nohmf"},
}

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
	playlist syncedPlaylist,
) (lcp.AppleMusicPlaylist, error) {
	playlistData, err := sendAppleMusicAPIRequest[playlistResponse](
		client,
		fmt.Sprintf("/v1/me/library/playlists/%s", playlist.AppleMusicID),
	)
	if err != nil {
		return lcp.AppleMusicPlaylist{}, fmt.Errorf(
			"fetching %s playlist: %w",
			playlist.AppleMusicID,
			err,
		)
	}

	var tracks []lcp.AppleMusicSong
	path := fmt.Sprintf("/v1/me/library/playlists/%s/tracks", playlist.AppleMusicID)
	for {
		trackData, err := sendAppleMusicAPIRequest[playlistTracksResponse](client, path)
		if err != nil {
			return lcp.AppleMusicPlaylist{}, fmt.Errorf(
				"fetching playlist data for %s: %w",
				path,
				err,
			)
		}
		for _, track := range trackData.Data {
			song, err := track.ToAppleMusicSong(client, rdb)
			if err != nil {
				return lcp.AppleMusicPlaylist{}, fmt.Errorf(
					"creating song from apple music song response: %w",
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

	duration := 0
	for _, track := range tracks {
		duration += track.DurationInMillis
	}

	return lcp.AppleMusicPlaylist{
		Name:             playlistData.Data[0].Attributes.Name,
		LastModified:     playlistData.Data[0].Attributes.LastModifiedDate,
		TrackCount:       len(tracks),
		DurationInMillis: duration,
		Tracks:           tracks,
		ID:               playlistData.Data[0].ID,
		URL: fmt.Sprintf(
			"https://music.apple.com/us/playlist/alt/%s",
			playlistData.Data[0].Attributes.PlayParams.GlobalID,
		),
		SpotifyID: playlist.SpotifyID,
	}, nil
}

func playlistEndpoint(c *cache.Cache[lcp.AppleMusicCache]) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth.SetCorsPolicy(w, r)

		id := r.PathValue("id")

		c.Mutex.RLock()
		defer c.Mutex.RUnlock()
		var p *lcp.AppleMusicPlaylist
		for _, playlist := range c.Data.Playlists {
			if playlist.ID == id {
				p = &playlist
				break
			}
		}

		if p == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		rawLast := r.URL.Query().Get("last")
		if rawLast != "" {
			n, err := strconv.Atoi(rawLast)
			if err != nil {
				http.Error(w, "invalid last", http.StatusBadRequest)
				return
			}
			n = min(n, 100)
			start := max(0, len(p.Tracks)-n)
			p.Tracks = p.Tracks[start:]

			resp := lcp.AppleMusicPlaylistResponse{
				Pagination: lcp.Pagination{Current: 1, Total: 1},
				Playlist:   *p,
			}
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(resp)
			if err != nil {
				err = fmt.Errorf("writing json to request: %w", err)
				util.InternalServerError(w, err, logAttr, "failed to encode json data")
			}
			return
		}

		var (
			page  = 1
			limit = 100
			total = int(math.Ceil(float64(len(p.Tracks)) / float64(limit)))
			next  *int
		)
		rawPage := r.URL.Query().Get("page")
		if rawPage != "" {
			n, err := strconv.Atoi(rawPage)
			if err != nil {
				http.Error(w, "invalid page", http.StatusBadRequest)
				return
			}
			page = n
		}
		if page > total {
			http.Error(w, "page doesn't exist", http.StatusBadRequest)
			return
		}
		start := min((page-1)*limit, len(p.Tracks))
		end := min(start+limit, len(p.Tracks))
		p.Tracks = p.Tracks[start:end]
		if page < total {
			nextPage := page + 1
			next = &nextPage
		}

		resp := lcp.AppleMusicPlaylistResponse{
			Pagination: lcp.Pagination{
				Current: page,
				Total:   total,
				Next:    next,
			},
			Playlist: *p,
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			err = fmt.Errorf("writing json to request: %w", err)
			util.InternalServerError(w, err, logAttr, "failed to encode json data")
		}
	})
}

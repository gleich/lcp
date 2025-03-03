package applemusic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp-2/internal/auth"
	"go.mattglei.ch/lcp-2/internal/cache"
	"go.mattglei.ch/lcp-2/pkg/lcp"
	"go.mattglei.ch/timber"
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

	var totalResponseData []songResponse
	trackData, err := sendAppleMusicAPIRequest[playlistTracksResponse](
		client,
		fmt.Sprintf("/v1/me/library/playlists/%s/tracks", id),
	)
	if err != nil {
		return lcp.AppleMusicPlaylist{}, err
	}
	totalResponseData = append(totalResponseData, trackData.Data...)
	for trackData.Next != "" {
		trackData, err = sendAppleMusicAPIRequest[playlistTracksResponse](client, trackData.Next)
		if err != nil {
			return lcp.AppleMusicPlaylist{}, fmt.Errorf(
				"%w failed to paginate through tracks for playlist with id of %s",
				err,
				id,
			)
		}
		totalResponseData = append(totalResponseData, trackData.Data...)
	}

	var tracks []lcp.AppleMusicSong
	for _, t := range totalResponseData {
		song, err := songFromSongResponse(client, rdb, t)
		if err != nil {
			return lcp.AppleMusicPlaylist{}, err
		}
		tracks = append(tracks, song)
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
	}, nil
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

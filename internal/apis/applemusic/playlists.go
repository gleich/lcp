package applemusic

import (
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/pkg/lcp"
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

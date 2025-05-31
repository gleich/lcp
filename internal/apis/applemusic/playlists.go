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

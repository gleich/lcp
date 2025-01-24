package applemusic

import (
	"fmt"
	"net/http"

	"pkg.mattglei.ch/lcp-2/pkg/models"
)

type recentlyPlayedResponse struct {
	Data []songResponse `json:"data"`
}

func fetchRecentlyPlayed(client *http.Client) ([]models.AppleMusicSong, error) {
	response, err := sendAppleMusicAPIRequest[recentlyPlayedResponse](
		client,
		"/v1/me/recent/played/tracks",
	)
	if err != nil {
		return []models.AppleMusicSong{}, err
	}

	var songs []models.AppleMusicSong
	for _, s := range response.Data {
		so, err := songFromSongResponse(s)
		if err != nil {
			return []models.AppleMusicSong{}, fmt.Errorf(
				"%v failed to parse song from song response",
				err,
			)
		}
		songs = append(songs, so)
	}

	// filter out duplicate songs
	seen := make(map[string]bool)
	uniqueSongs := []models.AppleMusicSong{}
	for _, song := range songs {
		if !seen[song.ID] {
			seen[song.ID] = true
			uniqueSongs = append(uniqueSongs, song)
		}
	}

	return uniqueSongs[:10], nil
}

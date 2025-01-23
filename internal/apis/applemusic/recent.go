package applemusic

import (
	"fmt"
	"net/http"
)

type recentlyPlayedResponse struct {
	Data []songResponse `json:"data"`
}

func fetchRecentlyPlayed(client *http.Client) ([]song, error) {
	response, err := sendAppleMusicAPIRequest[recentlyPlayedResponse](
		client,
		"/v1/me/recent/played/tracks",
	)
	if err != nil {
		return []song{}, err
	}

	var songs []song
	for _, s := range response.Data {
		so, err := songFromSongResponse(s)
		if err != nil {
			return []song{}, fmt.Errorf("%v failed to parse song from song response", err)
		}
		songs = append(songs, so)
	}

	// filter out duplicate songs
	seen := make(map[string]bool)
	uniqueSongs := []song{}
	for _, song := range songs {
		if !seen[song.ID] {
			seen[song.ID] = true
			uniqueSongs = append(uniqueSongs, song)
		}
	}

	return uniqueSongs[:10], nil
}

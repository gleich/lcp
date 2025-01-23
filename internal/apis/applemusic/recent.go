package applemusic

import "net/http"

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
		songs = append(songs, songFromSongResponse(s))
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

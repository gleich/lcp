package lcp

import "fmt"

// FetchPlaylist fetches a paginated page of tracks for the Apple Music playlist with the given id.
// Pages are 1-indexed and each page contains up to 100 tracks.
func FetchPlaylist(client *Client, id string, page int) (AppleMusicPlaylistResponse, error) {
	path := fmt.Sprintf("applemusic/playlists/%s?page=%d", id, page)
	resp, err := fetch[AppleMusicPlaylistResponse](client, path)
	if err != nil {
		return AppleMusicPlaylistResponse{}, fmt.Errorf("fetching playlist %s: %w", id, err)
	}
	return resp, nil
}

// FetchPlaylistLast fetches the last n tracks (up to 100) from the Apple Music playlist with the
// given id.
func FetchPlaylistLast(client *Client, id string, n int) (AppleMusicPlaylistResponse, error) {
	path := fmt.Sprintf("applemusic/playlists/%s?last=%d", id, n)
	resp, err := fetch[AppleMusicPlaylistResponse](client, path)
	if err != nil {
		return AppleMusicPlaylistResponse{}, fmt.Errorf("fetching last %d tracks from playlist %s: %w", n, id, err)
	}
	return resp, nil
}

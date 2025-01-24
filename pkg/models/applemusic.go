package models

import "time"

type AppleMusicCache struct {
	RecentlyPlayed []AppleMusicSong     `json:"recently_played"`
	Playlists      []AppleMusicPlaylist `json:"playlists"`
}

type AppleMusicSong struct {
	Track            string `json:"track"`
	Artist           string `json:"artist"`
	DurationInMillis int    `json:"duration_in_millis"`
	AlbumArtURL      string `json:"album_art_url"`
	URL              string `json:"url"`
	ID               string `json:"id"`
}

type AppleMusicPlaylist struct {
	Name         string           `json:"name"`
	Tracks       []AppleMusicSong `json:"tracks"`
	LastModified time.Time        `json:"last_modified"`
	URL          string           `json:"url"`
	ID           string           `json:"id"`
}

type AppleMusicPlaylistSummary struct {
	Name            string           `json:"name"`
	TrackCount      int              `json:"track_count"`
	FirstFourTracks []AppleMusicSong `json:"first_four_tracks"`
	ID              string           `json:"id"`
}

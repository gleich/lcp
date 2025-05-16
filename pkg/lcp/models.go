package lcp

import "time"

type CacheData interface {
	AppleMusicCache | []GitHubRepository | []SteamGame | []Workout
}

type AppleMusicCache struct {
	RecentlyPlayed []AppleMusicSong     `json:"recently_played"`
	Playlists      []AppleMusicPlaylist `json:"playlists"`
}

type AppleMusicSong struct {
	Track              string  `json:"track"`
	Artist             string  `json:"artist"`
	DurationInMillis   int     `json:"duration_in_millis"`
	AlbumArtURL        string  `json:"album_art_url"`
	AlbumArtPreviewURL string  `json:"album_art_preview_url"`
	AlbumArtBlurhash   string  `json:"album_art_blurhash"`
	URL                string  `json:"url"`
	ID                 string  `json:"id"`
	PreviewAudioURL    *string `json:"preview_audio_url"`
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

type GitHubRepository struct {
	Name          string    `json:"name"`
	Owner         string    `json:"owner"`
	Language      string    `json:"language"`
	LanguageColor string    `json:"language_color"`
	Description   string    `json:"description"`
	UpdatedAt     time.Time `json:"updated_at"`
	ID            string    `json:"id"`
	URL           string    `json:"url"`
}

type SteamGame struct {
	Name                string              `json:"name"`
	AppID               int32               `json:"app_id"`
	IconURL             string              `json:"icon_url"`
	RTimeLastPlayed     time.Time           `json:"rtime_last_played"`
	PlaytimeForever     int32               `json:"playtime_forever"`
	URL                 string              `json:"url"`
	HeaderURL           string              `json:"header_url"`
	LibraryHeroURL      string              `json:"library_hero_url"`
	LibraryHeroLogoURL  string              `json:"library_hero_logo_url"`
	AchievementProgress *float32            `json:"achievement_progress"`
	Achievements        *[]SteamAchievement `json:"achievements"`
}

type SteamAchievement struct {
	ApiName     string     `json:"api_name"`
	Achieved    bool       `json:"achieved"`
	Icon        string     `json:"icon"`
	DisplayName string     `json:"display_name"`
	Description *string    `json:"description"`
	UnlockTime  *time.Time `json:"unlock_time"`
}

type Workout struct {
	Platform           string         `json:"platform"`
	Name               string         `json:"name"`
	SportType          string         `json:"sport_type"`
	StartDate          time.Time      `json:"start_date"`
	MapBlurImage       *string        `json:"map_blur_image,omitempty"`
	MapImageURL        *string        `json:"map_image_url,omitempty"`
	HasMap             bool           `json:"has_map"`
	TotalElevationGain float32        `json:"total_elevation_gain,omitempty"`
	MovingTime         uint32         `json:"moving_time"`
	Distance           float32        `json:"distance,omitempty"`
	ID                 string         `json:"id"`
	HasHeartrate       bool           `json:"has_heartrate"`
	AverageHeartrate   float32        `json:"average_heartrate,omitempty"`
	HeartrateData      []int          `json:"heartrate_data"`
	HevyExercises      []HevyExercise `json:"hevy_exercises,omitempty"`
	HevyVolumeKG       float64        `json:"hevy_volume_kg,omitempty"`
	HevySetCount       int            `json:"hevy_set_count,omitempty"`
	Calories           float32        `json:"calories,omitempty"`
}

type HevyExercise struct {
	Title      string    `json:"title"`
	Sets       []HevySet `json:"sets"`
	SupersetID *int      `json:"superset_id"`
	ID         string    `json:"exercise_template_id"`
}

type HevySet struct {
	WeightKg float64 `json:"weight_kg"`
	Reps     int     `json:"reps"`
	Type     string  `json:"type"`
}

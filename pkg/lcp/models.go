package lcp

import "time"

type CacheData interface {
	AppleMusicCache | []GitHubRepository | []SteamGame | []StravaActivity | []HevyWorkout
}

type AppleMusicCache struct {
	RecentlyPlayed []AppleMusicSong     `json:"recently_played"`
	Playlists      []AppleMusicPlaylist `json:"playlists"`
}

type AppleMusicSong struct {
	Track            string  `json:"track"`
	Artist           string  `json:"artist"`
	DurationInMillis int     `json:"duration_in_millis"`
	AlbumArtURL      string  `json:"album_art_url"`
	AlbumArtBlurhash *string `json:"album_art_blurhash"`
	URL              string  `json:"url"`
	ID               string  `json:"id"`
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
	LibraryURL          *string             `json:"library_url"`
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

type StravaActivity struct {
	Name               string    `json:"name"`
	SportType          string    `json:"sport_type"`
	StartDate          time.Time `json:"start_date"`
	Timezone           string    `json:"timezone"`
	MapBlurImage       *string   `json:"map_blur_image"`
	MapImageURL        *string   `json:"map_image_url"`
	HasMap             bool      `json:"has_map"`
	TotalElevationGain float32   `json:"total_elevation_gain"`
	MovingTime         uint32    `json:"moving_time"`
	Distance           float32   `json:"distance"`
	ID                 uint64    `json:"id"`
	AverageHeartrate   float32   `json:"average_heartrate"`
	HeartrateData      []int     `json:"heartrate_data"`
	Calories           float32   `json:"calories"`
}

type HevyWorkout struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CreatedAt   time.Time      `json:"created_at"`
	Exercises   []HevyExercise `json:"exercises"`
}

type HevyExercise struct {
	Title string    `json:"title"`
	Sets  []HevySet `json:"sets"`
}

type HevySet struct {
	Index    int     `json:"index"`
	Type     string  `json:"type"`
	WeightKg float64 `json:"weight_kg"`
	Reps     int     `json:"reps"`
}

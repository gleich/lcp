package applemusic

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type song struct {
	Track            string `json:"track"`
	Artist           string `json:"artist"`
	DurationInMillis int    `json:"duration_in_millis"`
	AlbumArtURL      string `json:"album_art_url"`
	URL              string `json:"url"`
	ID               string `json:"id"`
}

type songResponse struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Href       string `json:"href"`
	Attributes struct {
		AlbumName        string   `json:"albumName"`
		GenreNames       []string `json:"genreNames"`
		TrackNumber      int      `json:"trackNumber"`
		ReleaseDate      string   `json:"releaseDate"`
		DurationInMillis int      `json:"durationInMillis"`
		Artwork          struct {
			Width  int    `json:"width"`
			Height int    `json:"height"`
			URL    string `json:"url"`
		} `json:"artwork"`
		URL        string `json:"url"`
		Name       string `json:"name"`
		ArtistName string `json:"artistName"`
		PlayParams struct {
			CatalogID string `json:"catalogId"`
		} `json:"playParams"`
	} `json:"attributes"`
}

func songFromSongResponse(s songResponse) (song, error) {
	if s.Attributes.URL == "" {
		// remove special characters
		slugURL := regexp.MustCompile(`[^\w\s-]`).ReplaceAllString(s.Attributes.Name, "")
		// replace spaces with hyphens
		slugURL = regexp.MustCompile(`\s+`).ReplaceAllString(slugURL, "-")

		u, err := url.JoinPath(
			"https://music.apple.com/us/song/",
			strings.ToLower(slugURL),
			fmt.Sprint(s.Attributes.PlayParams.CatalogID),
		)
		if err != nil {
			return song{}, fmt.Errorf("%v failed to create URL for song %s", err, s.Attributes.Name)
		}
		s.Attributes.URL = u
	}

	maxAlbumArtSize := 600.0
	return song{
		Track:            s.Attributes.Name,
		Artist:           s.Attributes.ArtistName,
		DurationInMillis: s.Attributes.DurationInMillis,
		AlbumArtURL: strings.ReplaceAll(strings.ReplaceAll(
			s.Attributes.Artwork.URL,
			"{w}",
			strconv.Itoa(int(math.Min(float64(s.Attributes.Artwork.Width), maxAlbumArtSize))),
		), "{h}", strconv.Itoa(int(math.Min(float64(s.Attributes.Artwork.Height), maxAlbumArtSize)))),
		URL: s.Attributes.URL,
		ID:  s.ID,
	}, nil
}

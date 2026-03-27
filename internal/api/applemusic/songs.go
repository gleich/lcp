package applemusic

import (
	"fmt"
	"image/jpeg"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/images"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

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
		Previews []struct {
			URL string `json:"url"`
		} `json:"previews"`
	} `json:"attributes"`
}

func (s songResponse) ToAppleMusicSong(
	client *http.Client,
	rdb *redis.Client,
) (lcp.AppleMusicSong, error) {
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
			return lcp.AppleMusicSong{}, fmt.Errorf(
				"failed to create URL for song \"%s\": %w",
				s.Attributes.Name,
				err,
			)
		}
		s.Attributes.URL = u
	}

	var (
		artURL           = albumArtURL(s, 400.0)
		albumArtBlurhash *string

		albumArtPermissionsExpiration *time.Time
	)
	id := s.ID
	if s.Attributes.PlayParams.CatalogID != "" {
		id = s.Attributes.PlayParams.CatalogID
	}
	if s.Attributes.Artwork.URL != "" {
		blurhash, err := images.BlurHash(client, rdb, *artURL, jpeg.Decode, logAttr)
		if err != nil && strings.Contains(err.Error(), "unexpected EOF") {
			timber.Warning(
				"unexpected EOF occurred while trying to create blur hash",
				timber.A("url", albumArtURL),
			)
		} else if err != nil {
			return lcp.AppleMusicSong{}, fmt.Errorf(
				"getting blur hash for \"%s\" (%s): %w",
				s.Attributes.Name,
				id,
				err,
			)
		}
		albumArtBlurhash = &blurhash

		artworkURL, err := url.Parse(*artURL)
		if err != nil {
			return lcp.AppleMusicSong{}, fmt.Errorf("parsing artwork url: %w", err)
		}
		const expirationKey = "X-Amz-Expires"
		query := artworkURL.Query()
		if query.Has(expirationKey) {
			expirationValue := query.Get(expirationKey)
			secs, err := strconv.Atoi(expirationValue)
			if err != nil {
				return lcp.AppleMusicSong{}, fmt.Errorf("parsing %s: %w", expirationValue, err)
			}
			expiration := time.Now().Add(time.Duration(secs) * time.Second)
			albumArtPermissionsExpiration = &expiration
		}
	}

	var previewAudioURL *string = nil
	if len(s.Attributes.Previews) > 0 {
		previewAudioURL = &s.Attributes.Previews[0].URL
	}

	return lcp.AppleMusicSong{
		Track:            s.Attributes.Name,
		Artist:           s.Attributes.ArtistName,
		DurationInMillis: s.Attributes.DurationInMillis,
		AlbumArtURL:      artURL,
		AlbumArtBlurhash: albumArtBlurhash,
		URL:              s.Attributes.URL,
		ID:               id,
		PreviewAudioURL:  previewAudioURL,

		AlbumArtPermissionsExpiration: albumArtPermissionsExpiration,
	}, nil
}

func albumArtURL(s songResponse, max float64) *string {
	if s.Attributes.Artwork.URL == "" {
		return nil
	}
	url := strings.ReplaceAll(strings.ReplaceAll(
		s.Attributes.Artwork.URL,
		"{w}",
		strconv.Itoa(int(math.Min(float64(s.Attributes.Artwork.Width), max))),
	), "{h}bb", strconv.Itoa(int(math.Min(float64(s.Attributes.Artwork.Height), max))))
	return &url
}

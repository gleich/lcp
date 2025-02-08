package applemusic

import (
	"context"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"pkg.mattglei.ch/lcp-2/internal/images"
	"pkg.mattglei.ch/lcp-2/pkg/lcp"
	"pkg.mattglei.ch/timber"
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
	} `json:"attributes"`
}

type BlurhashCacheEntry struct {
	Blurhash string    `json:"blurhash"`
	Created  time.Time `json:"created"`
	Url      string    `json:"url"`
}

func songFromSongResponse(
	client *http.Client,
	rdb *redis.Client,
	s songResponse,
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
				"%v failed to create URL for song %s",
				err,
				s.Attributes.Name,
			)
		}
		s.Attributes.URL = u
	}

	maxAlbumArtSize := 600.0
	albumArtURL := strings.ReplaceAll(strings.ReplaceAll(
		s.Attributes.Artwork.URL,
		"{w}",
		strconv.Itoa(int(math.Min(float64(s.Attributes.Artwork.Width), maxAlbumArtSize))),
	), "{h}", strconv.Itoa(int(math.Min(float64(s.Attributes.Artwork.Height), maxAlbumArtSize))))

	blurhash, err := loadAlbumArtBlurhash(client, rdb, albumArtURL, s.ID)
	if err != nil && strings.Contains(err.Error(), "unexpected EOF") {
		timber.Warning("failed to create blur hash for", albumArtURL)
	} else if err != nil {
		return lcp.AppleMusicSong{}, fmt.Errorf("%v failed to get blur hash for %s", err, s.ID)
	}

	return lcp.AppleMusicSong{
		Track:            s.Attributes.Name,
		Artist:           s.Attributes.ArtistName,
		DurationInMillis: s.Attributes.DurationInMillis,
		AlbumArtURL:      albumArtURL,
		AlbumArtBlurhash: blurhash,
		URL:              s.Attributes.URL,
		ID:               s.ID,
	}, nil
}

func loadAlbumArtBlurhash(
	client *http.Client,
	rdb *redis.Client,
	url string,
	id string,
) (*string, error) {
	ctx := context.Background()
	blurhashIsCached, err := rdb.Exists(ctx, id).Result()
	if err != nil {
		return nil, fmt.Errorf(
			"%v failed to check to see if id of %s from redis exists",
			err,
			id,
		)
	}

	if blurhashIsCached == 1 {
		var cachedBlurHash BlurhashCacheEntry
		result, err := rdb.Get(ctx, id).Result()
		if err != nil {
			return nil, fmt.Errorf(
				"%v failed to get blurhash for song with id of %s from redis",
				err,
				id,
			)
		}
		err = json.Unmarshal([]byte(result), &cachedBlurHash)
		if err != nil {
			return nil, fmt.Errorf("%v failed to decode json into blurhash cache entry", err)
		}
		return &cachedBlurHash.Blurhash, nil
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%v failed to create request for %s", err, url)
	}

	blurhashURL, err := createAlbumArtBlurhash(client, rdb, id, url, req)
	if err != nil {
		return nil, fmt.Errorf("%v failed to create blurhash", err)
	}

	return blurhashURL, nil
}

func updateAlbumArtPeriodically(client *http.Client, rdb *redis.Client) {
	for {
		time.Sleep(60 * time.Minute)
		var (
			cursor  uint64
			allKeys []string
			ctx     = context.Background()
		)
		for {
			keys, newCursor, err := rdb.Scan(context.Background(), cursor, "*", 100).Result()
			if err != nil {
				timber.Error(err, "failed to scan for keys from redis")
				return
			}
			allKeys = append(allKeys, keys...)
			if newCursor == 0 {
				break
			}
			cursor = newCursor
		}

		timber.Info(LOG_PREFIX, "checking album art blurhash for", len(allKeys), "albums")
		updated := 0
		for _, key := range allKeys {
			result, err := rdb.Get(ctx, key).Result()
			if err != nil {
				timber.Error(err, "failed to get key from redis", key)
				return
			}
			var cachedBlurHash BlurhashCacheEntry
			err = json.Unmarshal([]byte(result), &cachedBlurHash)
			if err != nil {
				timber.Error(err, "failed to decode json for key", key)
				return
			}

			req, err := http.NewRequest(http.MethodGet, cachedBlurHash.Url, nil)
			if err != nil {
				timber.Error(err, "failed to decode json")
			}
			req.Header.Set("If-Modified-Since", cachedBlurHash.Created.Format(time.RFC1123))

			updatedBlurhash, err := createAlbumArtBlurhash(
				client,
				rdb,
				key,
				cachedBlurHash.Url,
				req,
			)
			if err != nil && strings.Contains(err.Error(), "unexpected EOF") {
				timber.Warning("failed to create blur hash for", cachedBlurHash.Url)
			}
			if updatedBlurhash != nil && updatedBlurhash != &cachedBlurHash.Blurhash {
				cachedBlurHash.Blurhash = *updatedBlurhash
				updated++
			}
		}
		timber.Done("updated", fmt.Sprintf("%d/%d", updated, len(allKeys)), "album arts")
	}
}

func createAlbumArtBlurhash(
	client *http.Client,
	rdb *redis.Client,
	id string,
	url string,
	req *http.Request,
) (*string, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%v failed to execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified || resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%v failed to read response body from request", err)
	}

	blurhashURL, err := images.BlurImage(body, jpeg.Decode)
	if err != nil {
		return nil, fmt.Errorf("%v failed to blur image", err)
	}

	cacheData, err := json.Marshal(BlurhashCacheEntry{
		Blurhash: blurhashURL,
		Created:  time.Now(),
		Url:      url,
	})
	if err != nil {
		return nil, fmt.Errorf("%v failed to marshal cache data", err)
	}

	// approximately a 3 month long cache lifetime
	err = rdb.Set(context.Background(), id, string(cacheData), 2_191*time.Hour).
		Err()
	if err != nil {
		return nil, fmt.Errorf("%v failed to set %s to %s", err, id, string(cacheData))
	}
	timber.Done(LOG_PREFIX, "created blurhash for", id)
	return &blurhashURL, nil
}

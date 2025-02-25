package applemusic

import (
	"context"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp-2/internal/images"
	"go.mattglei.ch/timber"
)

type blurhashCache struct {
	Entires map[string]blurhashCacheEntry
	mutex   sync.RWMutex
}

type blurhashCacheEntry struct {
	Blurhash string
	Created  time.Time
	Url      string
}

// Load an album's album art blurhash either for the first time or from the cache.
// Instead of just fetching the image and creating the blurhash it checks the cache first.
//
// Returns:
//   - the URI encoded data for the blurhash or nil if there is no blurhash output
//   - error that might of been encountered
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
		var cachedBlurHash blurhashCacheEntry
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

// Update the album art in the cache every hour
func updateAlbumArtPeriodically(client *http.Client, rdb *redis.Client, interval time.Duration) {
	for {
		time.Sleep(interval)
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
			var cachedBlurHash blurhashCacheEntry
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

// Create an album art in the cache
//
// Returns:
//   - the URI encoded data for the blurhash or nil if there is no blurhash output
//   - error that might of been encountered
func createAlbumArtBlurhash(
	client *http.Client,
	rdb *redis.Client,
	id string,
	url string,
	req *http.Request,
) (*string, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w failed to execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified || resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w failed to read response body from request", err)
	}

	blurhashURL, err := images.BlurImage(body, jpeg.Decode)
	if err != nil {
		return nil, fmt.Errorf("%w failed to blur image", err)
	}

	cacheData, err := json.Marshal(blurhashCacheEntry{
		Blurhash: blurhashURL,
		Created:  time.Now(),
		Url:      url,
	})
	if err != nil {
		return nil, fmt.Errorf("%v failed to marshal cache data", err)
	}

	// approximately a 1 week long cache lifetime
	err = rdb.Set(context.Background(), id, string(cacheData), 168*time.Hour).
		Err()
	if err != nil {
		return nil, fmt.Errorf("%v failed to set %s to %s", err, id, string(cacheData))
	}
	return &blurhashURL, nil
}

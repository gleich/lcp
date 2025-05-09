package applemusic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/apis"
	"go.mattglei.ch/lcp/internal/images"
	"go.mattglei.ch/timber"
)

type blurhashCacheEntry struct {
	Blurhash string
	Created  time.Time
	Url      string
}

// loadAlbumArtBlurhash retrieves the blurhash for album art by first checking
// the Redis cache. If a cached blurhash is found (using the given id), it is returned.
// Otherwise, the album art is fetched from the provided URL, a new blurhash is created,
// and its URI-encoded value is returned.
//
// If no blurhash can be generated, nil is returned. Any errors encountered during
// the cache lookup, HTTP request, or processing are returned.
func loadAlbumArtBlurhash(
	client *http.Client,
	rdb *redis.Client,
	url string,
	id string,
) (*string, error) {
	if strings.Contains(url, "blobstore.apple.com") {
		return nil, nil
	}

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
	if err != nil && !errors.Is(err, apis.WarningError) {
		return nil, fmt.Errorf("%v failed to create blurhash", err)
	}

	return blurhashURL, nil
}

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

		timber.Info(logPrefix, "checking album art blurhash for", len(allKeys), "albums")
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

			if strings.Contains(cachedBlurHash.Url, "blobstore.apple.com") {
				continue
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
			if err != nil && !errors.Is(err, apis.WarningError) {
				timber.Error(err, "failed to generate blur hash for", key)
			}
			if updatedBlurhash != nil && updatedBlurhash != &cachedBlurHash.Blurhash {
				cachedBlurHash.Blurhash = *updatedBlurhash
				updated++
			}
		}
		timber.Done("updated", fmt.Sprintf("%d/%d", updated, len(allKeys)), "album arts")
	}
}

// createAlbumArtBlurhash generates a blurhash from the album art image fetched via
// the provided HTTP request. It caches the resulting URI-encoded blurhash in Redis
// under the given id and returns it. If the HTTP response indicates that the image
// is unchanged or not found, the function returns nil. Any errors encountered during
// the request, image processing, or caching are returned.
func createAlbumArtBlurhash(
	client *http.Client,
	rdb *redis.Client,
	id string,
	url string,
	req *http.Request,
) (*string, error) {
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
	)
	body, err := apis.Request(logPrefix, client, req)
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

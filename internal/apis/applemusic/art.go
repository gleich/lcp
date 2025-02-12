package applemusic

import (
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

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
	cache *blurhashCache,
	url string,
	id string,
) (*string, error) {
	cache.mutex.RLock()
	cachedBlurhash, exists := cache.Entires[id]
	cache.mutex.RUnlock()

	if exists {
		blurHashCopy := cachedBlurhash.Blurhash
		return &blurHashCopy, nil
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%v failed to create request for %s", err, url)
	}

	blurhashURL, err := createAlbumArtBlurhash(client, cache, id, url, req)
	if err != nil {
		return nil, fmt.Errorf("%v failed to create blurhash", err)
	}

	return blurhashURL, nil
}

// Update the album art in the cache every hour
func updateAlbumArtPeriodically(client *http.Client, cache *blurhashCache, interval time.Duration) {
	for {
		time.Sleep(interval)

		timber.Info(LOG_PREFIX, "checking album art blurhash for", len(cache.Entires), "albums")
		updated := 0

		cache.mutex.RLock()
		entriesCopy := make(map[string]blurhashCacheEntry, len(cache.Entires))
		for id, entry := range cache.Entires {
			entriesCopy[id] = entry
		}
		cache.mutex.RUnlock()

		for id, entry := range entriesCopy {
			req, err := http.NewRequest(http.MethodGet, entry.Url, nil)
			if err != nil {
				timber.Error(err, "failed to decode json")
			}
			req.Header.Set("If-Modified-Since", entry.Created.Format(time.RFC1123))

			updatedBlurhash, err := createAlbumArtBlurhash(
				client,
				cache,
				id,
				entry.Url,
				req,
			)
			if err != nil && strings.Contains(err.Error(), "unexpected EOF") {
				timber.Warning("failed to create blur hash for", entry.Url)
			}
			if updatedBlurhash != nil {
				cache.mutex.Lock()
				currentEntry := cache.Entires[id]
				cache.mutex.Unlock()
				if *updatedBlurhash != currentEntry.Blurhash {
					// The cache update already happens inside createAlbumArtBlurhash,
					// so this may be just for counting purposes.
					updated++
				}
			}
		}
		timber.Done("updated", fmt.Sprintf("%d/%d", updated, len(cache.Entires)), "album arts")
	}
}

// Create an album art in the cache
//
// Returns:
//   - the URI encoded data for the blurhash or nil if there is no blurhash output
//   - error that might of been encountered
func createAlbumArtBlurhash(
	client *http.Client,
	cache *blurhashCache,
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

	cache.mutex.Lock()
	cache.Entires[id] = blurhashCacheEntry{
		Blurhash: blurhashURL,
		Created:  time.Now(),
		Url:      url,
	}
	cache.mutex.Unlock()
	return &blurhashURL, nil
}

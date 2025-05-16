package images

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/apis"
)

type cacheEntry struct {
	BlurHash string
	Created  time.Time
	URL      string
}

// BlurHash looks up or generates a BlurHash for url, caching the result in Redis and returning the
// hash.
func BlurHash(
	client *http.Client,
	rdb *redis.Client,
	url string,
	decoder ImageDecoder,
) (string, error) {
	ctx := context.Background()
	result, err := rdb.Get(ctx, url).Result()
	if err == redis.Nil {
		blurhash, err := createCacheEntry(client, rdb, url, decoder)
		if err != nil {
			return "", fmt.Errorf("%w failed to generate blurhash for %s", err, url)
		}
		return blurhash, nil
	} else if err != nil {
		return "", fmt.Errorf("%w failed to get %s from redis cache", err, url)
	}

	var cachedBlurhash *cacheEntry
	err = json.Unmarshal([]byte(result), &cachedBlurhash)
	if err != nil {
		return "", fmt.Errorf("%w failed to parse JSON for blurhash from \"%s\"", err, result)
	}
	return cachedBlurhash.BlurHash, nil
}

// createCacheEntry downloads an image, computes its BlurHash, stores it in Redis for one week, and
// returns the hash.
func createCacheEntry(
	client *http.Client,
	rdb *redis.Client,
	url string,
	decoder ImageDecoder,
) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("%w failed to create request for %s", err, url)
	}
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
	)

	body, err := apis.Request("[image cache]", client, req)
	if err != nil {
		return "", fmt.Errorf("%w failed to read response body from request", err)
	}

	blurhash, err := blur(body, decoder)
	if err != nil {
		return "", fmt.Errorf("%w failed to blur image", err)
	}

	cacheData, err := json.Marshal(cacheEntry{
		BlurHash: blurhash,
		Created:  time.Now(),
		URL:      url,
	})
	if err != nil {
		return "", fmt.Errorf("%v failed to marshal cache data", err)
	}

	// approximately a 1 week long cache lifetime
	err = rdb.Set(context.Background(), url, string(cacheData), 168*time.Hour).
		Err()
	if err != nil {
		return "", fmt.Errorf("%v failed to set %s to %s", err, url, string(cacheData))
	}
	return blurhash, nil
}

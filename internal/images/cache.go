package images

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/timber"
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
		blurhash, err := createCacheEntry(client, rdb, url, decoder, ctx)
		if err != nil {
			return "", fmt.Errorf("generating blurhash for %s: %w", url, err)
		}
		return blurhash, nil
	} else if err != nil {
		return "", fmt.Errorf("getting %s from redis cache: %w", url, err)
	}

	var cachedBlurhash *cacheEntry
	err = json.Unmarshal([]byte(result), &cachedBlurhash)
	if err != nil {
		timber.Debug(result)
		return "", fmt.Errorf("parsing json for blurhash: %w", err)
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
	ctx context.Context,
) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
	)

	body, err := api.Request("[image cache]", client, req)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	blurhash, err := blur(body, decoder)
	if err != nil {
		return "", fmt.Errorf("blurring image: %w", err)
	}

	cacheData, err := json.Marshal(cacheEntry{
		BlurHash: blurhash,
		Created:  time.Now(),
		URL:      url,
	})
	if err != nil {
		return "", fmt.Errorf("marshaling cache json data: %w", err)
	}

	// approximately a 1 week long cache lifetime
	err = rdb.Set(ctx, url, string(cacheData), 168*time.Hour).
		Err()
	if err != nil {
		return "", fmt.Errorf("setting value for %s in redis: %w", url, err)
	}
	return blurhash, nil
}

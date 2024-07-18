package cache

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gleich/lcp/pkg/secrets"
	"github.com/gleich/lumber/v2"
)

type Cache[T any] struct {
	mutex sync.RWMutex
	data  T
}

func New[T any](data T) Cache[T] {
	return Cache[T]{
		data: data,
	}
}

// Handle a GET request to load data from the given cache
func Route[T any](cache *Cache[T], loadedSecrets secrets.Secrets) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+loadedSecrets.ValidToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		cache.mutex.RLock()
		defer cache.mutex.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(cache.data)
		if err != nil {
			lumber.Error(err, "Failed to write data")
		}
	})
}

// Update the given cache
func Update[T any](cache *Cache[T], data T) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.data = data
}

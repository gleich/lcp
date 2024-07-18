package cache

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gleich/lcp/pkg/secrets"
	"github.com/gleich/lumber/v2"
)

type Cache[T any] struct {
	Name    string
	mutex   sync.RWMutex
	data    T
	updated time.Time
}

func New[T any](name string, data T) Cache[T] {
	return Cache[T]{
		Name:    name,
		data:    data,
		updated: time.Now(),
	}
}

type response[T any] struct {
	Data    T         `json:"data"`
	Updated time.Time `json:"updated"`
}

// Handle a GET request to load data from the given cache
func (c *Cache[T]) Route(loadedSecrets secrets.Secrets) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+loadedSecrets.ValidToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		c.mutex.RLock()
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response[T]{Data: c.data, Updated: c.updated})
		c.mutex.RUnlock()
		if err != nil {
			lumber.Error(err, "failed to write data")
		}
	})
}

// Update the given cache
func (c *Cache[T]) Update(data T) {
	var updated bool
	c.mutex.Lock()
	if !reflect.DeepEqual(data, c.data) {
		c.data = data
		c.updated = time.Now()
	}
	c.mutex.Unlock()
	if updated {
		lumber.Success(c.Name, "updated")
	}
}

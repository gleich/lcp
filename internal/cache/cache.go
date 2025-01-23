package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"pkg.mattglei.ch/lcp-2/internal/apis"
	"pkg.mattglei.ch/lcp-2/internal/auth"
	"pkg.mattglei.ch/lcp-2/internal/secrets"
	"pkg.mattglei.ch/timber"
)

type Cache[T any] struct {
	name      string
	DataMutex sync.RWMutex
	Data      T
	Updated   time.Time
	filePath  string
}

func New[T any](name string, data T, update bool) *Cache[T] {
	cache := Cache[T]{
		name:     name,
		Updated:  time.Now(),
		filePath: filepath.Join(secrets.ENV.CacheFolder, fmt.Sprintf("%s.json", name)),
	}
	cache.loadFromFile()
	if update {
		cache.Update(data)
	}
	return &cache
}

type CacheResponse[T any] struct {
	Data    T         `json:"data"`
	Updated time.Time `json:"updated"`
}

func (c *Cache[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAuthorized(w, r) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	c.DataMutex.RLock()
	err := json.NewEncoder(w).Encode(CacheResponse[T]{Data: c.Data, Updated: c.Updated})
	c.DataMutex.RUnlock()
	if err != nil {
		err = fmt.Errorf("%v failed to write json data to request", err)
		timber.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Cache[T]) Update(data T) {
	c.DataMutex.RLock()
	old, err := json.Marshal(c.Data)
	if err != nil {
		timber.Error(err, "failed to json marshal old data")
		return
	}
	c.DataMutex.RUnlock()
	new, err := json.Marshal(data)
	if err != nil {
		timber.Error(err, "failed to json marshal new data")
		return
	}

	if string(old) != string(new) && string(new) != "null" && strings.Trim(string(new), " ") != "" {
		c.DataMutex.Lock()
		c.Data = data
		c.Updated = time.Now()
		c.DataMutex.Unlock()

		c.persistToFile()
		timber.Done(strings.ToUpper(c.name), "cache updated")
	}
}

func UpdatePeriodically[T any, C any](
	cache *Cache[T],
	client C,
	update func(C) (T, error),
	interval time.Duration,
) {
	for {
		time.Sleep(interval)
		data, err := update(client)
		if err != nil {
			if !errors.Is(err, apis.WarningError) {
				timber.Error(err, "updating", cache.name, "cache failed")
			}
		} else {
			cache.Update(data)
		}
	}
}

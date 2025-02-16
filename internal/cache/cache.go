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

	"go.mattglei.ch/lcp-2/internal/apis"
	"go.mattglei.ch/lcp-2/internal/auth"
	"go.mattglei.ch/lcp-2/internal/secrets"
	"go.mattglei.ch/lcp-2/pkg/lcp"
	"go.mattglei.ch/timber"
)

type Cache[T lcp.CacheData] struct {
	name      string
	DataMutex sync.RWMutex
	Data      T
	Updated   time.Time
	filePath  string
}

func New[T lcp.CacheData](name string, data T, update bool) *Cache[T] {
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
		err = fmt.Errorf("%w failed to write json data to request", err)
		timber.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Cache[T]) Update(data T) {
	const jsonIndent = "  "
	c.DataMutex.RLock()
	oldBin, err := json.MarshalIndent(c.Data, "", jsonIndent)
	if err != nil {
		timber.Error(err, "failed to json marshal old data")
		return
	}
	c.DataMutex.RUnlock()
	newBin, err := json.MarshalIndent(data, "", jsonIndent)
	if err != nil {
		timber.Error(err, "failed to json marshal new data")
		return
	}

	old := string(oldBin)
	new := string(newBin)
	if old != new && new != "null" && strings.Trim(string(newBin), " ") != "" {
		c.DataMutex.Lock()
		c.Data = data
		c.Updated = time.Now()
		c.DataMutex.Unlock()

		c.persistToFile()
		timber.Done(fmt.Sprintf("[%s]", c.name), "cache updated")
	}
}

func UpdatePeriodically[T lcp.CacheData, C any](
	cache *Cache[T],
	client C,
	update func(C) (T, error),
	interval time.Duration,
) {
	for {
		time.Sleep(interval)
		data, err := update(client)
		if err != nil {
			if !errors.Is(err, apis.IgnoreError) {
				timber.Error(err, "updating", cache.name, "cache failed")
			}
		} else {
			cache.Update(data)
		}
	}
}

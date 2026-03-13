package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/internal/tasks"
	"go.mattglei.ch/lcp/pkg/lcp"
)

type CacheInstance int

const (
	AppleMusic CacheInstance = iota
	Workouts
	GitHub
	Steam
)

func (c CacheInstance) String() string {
	switch c {
	case AppleMusic:
		return "applemusic"
	case Workouts:
		return "workouts"
	case GitHub:
		return "github"
	case Steam:
		return "steam"
	}
	return "unknown"
}

func (c CacheInstance) LogPrefix() string {
	return fmt.Sprintf("[%s]", c.String())
}

type Cache[T lcp.CacheData] struct {
	instance CacheInstance
	filePath string
	logCtx   []any

	Mutex   sync.RWMutex
	Data    T
	Updated time.Time

	// Diff function to to check to see if the cache should be updated because the data has changed
	Diff func(c *Cache[T], new, old T) (bool, error)

	// Custom JSON marshalling for the endpoint. Used to control what content actually gets returned
	// from the cache via the endpoint.
	MarshalResponse func(c *Cache[T]) ([]byte, error)

	connections      map[chan string]struct{}
	connectionsMutex sync.Mutex
}

func New[T lcp.CacheData](instance CacheInstance, data T, update bool) *Cache[T] {
	start := time.Now()
	cache := Cache[T]{
		instance: instance,
		Updated:  time.Now().UTC(),
		filePath: filepath.Join(
			secrets.ENV.CacheFolder,
			fmt.Sprintf("%s.son", instance.String()),
		),
		logCtx:      []any{"cache", instance.String()},
		connections: make(map[chan string]struct{}),
		MarshalResponse: func(c *Cache[T]) ([]byte, error) {
			data, err := json.Marshal(lcp.CacheResponse[T]{Data: c.Data, Updated: c.Updated})
			if err != nil {
				return []byte{}, fmt.Errorf("encoding json data: %w", err)
			}
			return data, nil
		},
		Diff: func(c *Cache[T], new, old T) (bool, error) {
			oldBin, err := json.Marshal(old)
			if err != nil {
				return false, fmt.Errorf("marshal old json: %w", err)
			}
			newBin, err := json.Marshal(new)
			if err != nil {
				return false, fmt.Errorf("marshal new json: %w", err)
			}
			var (
				newJSON = string(newBin)
				oldJSON = string(oldBin)
			)

			return oldJSON != newJSON && newJSON != "null" && strings.Trim(newJSON, " ") != "", nil
		},
	}
	cache.loadFromFile()
	if update {
		cache.Update(start, data)
	}
	return &cache
}

func (c *Cache[T]) Update(start time.Time, data T) {
	task := tasks.Cache.Update
	c.Mutex.RLock()
	changed, err := c.Diff(c, data, c.Data)
	if err != nil {
		task.Error(err, "checking for diff between old and new elements", c.logCtx...)
		return
	}
	c.Mutex.RUnlock()
	if changed {
		c.Mutex.Lock()
		c.Data = data
		c.Updated = time.Now().UTC()
		c.Mutex.Unlock()

		c.persistToFile()
		task.InfoSince("updated", start, c.logCtx...)

		if len(c.connections) != 0 {
			task = tasks.Cache.ServeStream
			start = time.Now()
			// broadcast update to connections
			c.Mutex.RLock()
			frame, err := c.MarshalResponse(c)
			if err != nil {
				task.Error(err, "failed to create endpoint data", c.logCtx...)
				return
			}
			c.Mutex.RUnlock()

			c.connectionsMutex.Lock()
			for connection := range c.connections {
				select {
				case connection <- string(frame):
				default:
					delete(c.connections, connection)
					close(connection)
				}
			}
			c.connectionsMutex.Unlock()
			task.InfoSince(
				fmt.Sprintf("updated %d connection(s)", len(c.connections)),
				start,
				c.logCtx...)
		}
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
		start := time.Now()
		data, err := update(client)
		if err != nil {
			if !slices.ContainsFunc(
				ExpectedErrors,
				func(e error) bool { return errors.Is(err, e) },
			) {
				tasks.Cache.Update.Error(err, "failed to update", cache.logCtx...)
			}
		} else {
			cache.Update(start, data)
		}
	}
}

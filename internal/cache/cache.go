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
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
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

	Mutex   sync.RWMutex
	Data    T
	Updated time.Time

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
			fmt.Sprintf("%s.json", instance.String()),
		),
		connections: make(map[chan string]struct{}),
		MarshalResponse: func(c *Cache[T]) ([]byte, error) {
			data, err := json.Marshal(lcp.CacheResponse[T]{Data: c.Data, Updated: c.Updated})
			if err != nil {
				return []byte{}, fmt.Errorf("encoding json data: %w", err)
			}
			return data, nil
		},
	}
	cache.loadFromFile()
	if update {
		cache.Update(start, data)
	}
	return &cache
}

func (c *Cache[T]) Update(start time.Time, data T) {
	c.Mutex.RLock()
	oldBin, err := json.Marshal(c.Data)
	if err != nil {
		timber.Error(err, "failed to json marshal old data")
		return
	}
	c.Mutex.RUnlock()
	newBin, err := json.Marshal(data)
	if err != nil {
		timber.Error(err, "failed to json marshal new data")
		return
	}

	new := string(newBin)
	old := string(oldBin)
	if old != new && new != "null" && strings.Trim(new, " ") != "" {
		// oldFmt, err := json.MarshalIndent(c.Data, "", "  ")
		// if err != nil {
		// 	timber.Error(err, "failed to format old json data")
		// 	return
		// }
		// newFmt, err := json.MarshalIndent(data, "", "  ")
		// if err != nil {
		// 	timber.Error(err, "failed to format new json data")
		// 	return
		// }
		// os.WriteFile(fmt.Sprintf("%s-old.json", c.instance.LogPrefix()), oldFmt, 0655)
		// os.WriteFile(fmt.Sprintf("%s-new.json", c.instance.LogPrefix()), newFmt, 0655)

		c.Mutex.Lock()
		c.Data = data
		c.Updated = time.Now().UTC()
		c.Mutex.Unlock()

		c.persistToFile()
		timber.DoneSince(start, c.instance.LogPrefix(), "cache updated")

		if len(c.connections) != 0 {
			start = time.Now()
			// broadcast update to connections
			c.Mutex.RLock()
			frame, err := c.MarshalResponse(c)
			if err != nil {
				timber.Error(err, "failed to create endpoint data")
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

			connWord := "connection"
			if len(c.connections) > 1 {
				connWord = "connections"
			}
			timber.DoneSince(start, c.instance.LogPrefix(), "updated", len(c.connections), connWord)
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
				timber.Error(err, cache.instance.LogPrefix(), "updating cache failed")
			}
		} else {
			cache.Update(start, data)
		}
	}
}

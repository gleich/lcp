package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.mattglei.ch/lcp/internal/apis"
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

	MarshalResponse func(c *Cache[T]) (string, error)

	connections      map[chan string]struct{}
	connectionsMutex sync.Mutex
}

func New[T lcp.CacheData](instance CacheInstance, data T, update bool) *Cache[T] {
	cache := Cache[T]{
		instance: instance,
		Updated:  time.Now().UTC(),
		filePath: filepath.Join(
			secrets.ENV.CacheFolder,
			fmt.Sprintf("%s.json", instance.String()),
		),
		connections: make(map[chan string]struct{}),
		MarshalResponse: func(c *Cache[T]) (string, error) {
			data, err := json.Marshal(lcp.CacheResponse[T]{Data: c.Data, Updated: c.Updated})
			if err != nil {
				return "", fmt.Errorf("%w failed to encode json data", err)
			}
			return string(data), nil
		},
	}
	cache.loadFromFile()
	if update {
		cache.Update(data)
	}
	return &cache
}

func (c *Cache[T]) Update(data T) {
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
	if string(oldBin) != new && new != "null" && strings.Trim(new, " ") != "" {
		c.Mutex.Lock()
		c.Data = data
		c.Updated = time.Now().UTC()
		c.Mutex.Unlock()

		c.persistToFile()
		timber.Done(c.instance.LogPrefix(), "cache updated")

		if len(c.connections) != 0 {
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
				case connection <- frame:
				default:
					delete(c.connections, connection)
					close(connection)
				}
			}
			c.connectionsMutex.Unlock()
			if len(c.connections) > 1 {
				timber.Done(c.instance.LogPrefix(), "updated", len(c.connections), "connections")
			} else {
				timber.Done(c.instance.LogPrefix(), "updated", len(c.connections), "connection")
			}
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
		data, err := update(client)
		if err != nil {
			if !errors.Is(err, apis.ErrWarning) && !errors.Is(err, ErrAppleMusicNoArtwork) &&
				!errors.Is(err, ErrSteamOwnedGamesEmpty) {
				timber.Error(err, cache.instance.LogPrefix(), "updating cache failed")
			}
		} else {
			cache.Update(data)
		}
	}
}

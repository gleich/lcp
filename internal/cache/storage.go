package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"go.mattglei.ch/lcp/pkg/lcp"
)

func (c *Cache[T]) persistToFile() {
	c.Mutex.RLock()
	bin, err := json.Marshal(lcp.CacheResponse[T]{Data: c.Data, Updated: c.Updated})
	c.Mutex.RUnlock()
	if err != nil {
		c.Logger.Error().Err(err).Msg("encoding data to json failed")
		return
	}
	err = os.MkdirAll(filepath.Dir(c.filePath), 0755)
	if err != nil {
		c.Logger.Error().Err(err).Str("path", c.filePath).Msg("creating cache directory failed")
		return
	}
	err = os.WriteFile(c.filePath, bin, 0666)
	if err != nil {
		c.Logger.Error().Err(err).Str("path", c.filePath).Msg("writing cache file failed")
	}
}

func (c *Cache[T]) loadFromFile() {
	if _, err := os.Stat(c.filePath); !os.IsNotExist(err) {
		b, err := os.ReadFile(c.filePath)
		if err != nil {
			c.Logger.Fatal().Err(err).Str("path", c.filePath).Msg("reading from cache file failed")
		}

		var data lcp.CacheResponse[T]
		err = json.Unmarshal(b, &data)
		if err != nil {
			c.Logger.Fatal().Err(err).Str("path", c.filePath).Msg("unmarshaling json data failed")
		}

		c.Data = data.Data
		c.Updated = data.Updated
	}
}

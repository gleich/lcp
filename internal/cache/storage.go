package cache

import (
	"encoding/json"
	"os"

	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

func (c *Cache[T]) persistToFile() {
	c.Mutex.RLock()
	bin, err := json.Marshal(lcp.CacheResponse[T]{Data: c.Data, Updated: c.Updated})
	c.Mutex.RUnlock()
	if err != nil {
		timber.Error(err, "encoding data to json failed", c.LogAttr)
		return
	}
	if err = os.WriteFile(c.filePath, bin, 0666); err != nil {
		timber.Error(err, "writing cache file failed", timber.A("path", c.filePath), c.LogAttr)
	}
}

func (c *Cache[T]) loadFromFile() {
	if _, err := os.Stat(c.filePath); !os.IsNotExist(err) {
		b, err := os.ReadFile(c.filePath)
		if err != nil {
			timber.Fatal(
				err,
				"reading from cache file failed",
				timber.A("path", c.filePath),
				c.LogAttr,
			)
		}

		var data lcp.CacheResponse[T]
		err = json.Unmarshal(b, &data)
		if err != nil {
			timber.Fatal(
				err,
				"unmarshaling json data failed",
				timber.A("path", c.filePath),
				c.LogAttr,
			)
		}

		c.Data = data.Data
		c.Updated = data.Updated
	}
}

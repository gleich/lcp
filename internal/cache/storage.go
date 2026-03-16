package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

func (c *Cache[T]) persistToFile() {
	var file *os.File
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	if _, err := os.Stat(c.filePath); os.IsNotExist(err) {
		folder := filepath.Dir(c.filePath)
		err := os.MkdirAll(folder, 0700)
		if err != nil {
			timber.Error(err, "failed to create folder", timber.A("path", folder), c.LogAttr)
			return
		}
		file, err = os.Create(c.filePath)
		if err != nil {
			timber.Error(err, "failed to create file", timber.A("path", c.filePath), c.LogAttr)
			return
		}
	} else {
		file, err = os.OpenFile(c.filePath, os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			timber.Error(err, "failed to read file", timber.A("path", c.filePath), c.LogAttr)
			return
		}
	}

	c.Mutex.RLock()
	bin, err := json.Marshal(lcp.CacheResponse[T]{
		Data:    c.Data,
		Updated: c.Updated,
	})
	c.Mutex.RUnlock()
	if err != nil {
		timber.Error(err, "encoding data to json failed", c.LogAttr)
		return
	}
	_, err = file.Write(bin)
	if err != nil {
		timber.Error(err, "writing data to json failed", c.LogAttr)
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

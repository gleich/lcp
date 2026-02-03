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
	defer func() { _ = file.Close() }()
	if _, err := os.Stat(c.filePath); os.IsNotExist(err) {
		folder := filepath.Dir(c.filePath)
		err := os.MkdirAll(folder, 0700)
		if err != nil {
			timber.Error(err, "failed to create folder at path:", folder)
			return
		}
		file, err = os.Create(c.filePath)
		if err != nil {
			timber.Error(err, "failed to create file at path:", c.filePath)
			return
		}
	} else {
		file, err = os.OpenFile(c.filePath, os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			timber.Error(err, "failed to read file at path:", c.filePath)
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
		timber.Error(err, "encoding data to json failed")
		return
	}
	_, err = file.Write(bin)
	if err != nil {
		timber.Error(err, "writing data to json failed")
	}

}

func (c *Cache[T]) loadFromFile() {
	if _, err := os.Stat(c.filePath); !os.IsNotExist(err) {
		b, err := os.ReadFile(c.filePath)
		if err != nil {
			timber.Fatal(err, "reading from cache file from", c.filePath, "failed")
		}

		var data lcp.CacheResponse[T]
		err = json.Unmarshal(b, &data)
		if err != nil {
			timber.Fatal(err, "unmarshaling json data from", c.filePath, "failed:", string(b))
		}

		c.Data = data.Data
		c.Updated = data.Updated
	}
}

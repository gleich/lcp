package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"pkg.mattglei.ch/timber"
)

func (c *Cache[T]) persistToFile() {
	var file *os.File
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
	defer file.Close()

	c.DataMutex.RLock()
	b, err := json.Marshal(CacheResponse[T]{
		Data:    c.Data,
		Updated: c.Updated,
	})
	c.DataMutex.RUnlock()
	if err != nil {
		timber.Error(err, "encoding data to json failed")
		return
	}
	_, err = file.Write(b)
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

		var data CacheResponse[T]
		err = json.Unmarshal(b, &data)
		if err != nil {
			timber.Fatal(err, "unmarshal json data failed from:", string(b))
		}

		c.Data = data.Data
		c.Updated = data.Updated
	}
}

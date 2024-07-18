package cache

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type Cache struct {
	mutex sync.Mutex
	data  gin.H
}

func New() Cache {
	return Cache{
		mutex: sync.Mutex{},
		data:  gin.H{},
	}
}

func CacheRoute(cache *Cache, c *gin.Context) {
	cache.mutex.Lock()
	c.JSON(http.StatusOK, cache.data)
	cache.mutex.Unlock()
}

func Update(cache *Cache, data gin.H) {
	cache.mutex.Lock()
	cache.data = data
	cache.mutex.Unlock()
}

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gleich/lcp/pkg/cache"
	"github.com/gleich/lumber/v2"
)

func main() {
	lumber.Info("booted")

	// caches
	stravaCache := cache.New()

	r := gin.Default()
	r.GET("/", rootRedirect)
	r.GET("/strava/cache", func(ctx *gin.Context) { cache.CacheRoute(&stravaCache, ctx) })

	err := r.Run()
	if err != nil {
		lumber.Fatal(err, "running gin failed")
	}
}

func rootRedirect(ctx *gin.Context) {
	ctx.Redirect(http.StatusTemporaryRedirect, "https://mattglei.ch")
}

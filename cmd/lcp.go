package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/api/applemusic"
	"go.mattglei.ch/lcp/internal/api/github"
	"go.mattglei.ch/lcp/internal/api/steam"
	"go.mattglei.ch/lcp/internal/api/workouts"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/middleware"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/timber"
)

func main() {
	start := time.Now()
	secrets.Load()
	if secrets.ENV.StructuredLogging {
		timber.Structured(true)
	} else {
		ny, err := time.LoadLocation("America/New_York")
		if err != nil {
			timber.Fatal(err, "failed to load new york timezone")
		}
		timber.Timezone(ny)
		timber.TimeFormat("01/02 03:04:05 PM MST")
	}

	timber.InfoSince(start, "booted")

	var (
		client = api.IPV4OnlyClient()
		mux    = http.NewServeMux()
		rdb    = redis.NewClient(&redis.Options{
			Addr:     secrets.ENV.RedisAddress,
			Password: secrets.ENV.RedisPassword,
			DB:       0,
			MaintNotificationsConfig: &maintnotifications.Config{
				Mode: maintnotifications.ModeDisabled,
			},
		})
	)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://mattglei.ch/writing/lcp", http.StatusPermanentRedirect)
	})

	setups := map[cache.CacheInstance]func(){
		cache.GitHub:     func() { github.Setup(mux) },
		cache.Workouts:   func() { workouts.Setup(mux, client, rdb) },
		cache.Steam:      func() { steam.Setup(mux, client, rdb) },
		cache.AppleMusic: func() { applemusic.Setup(mux, client, rdb) },
	}
	var wg sync.WaitGroup
	for cacheInstance, setup := range setups {
		wg.Go(func() {
			start := time.Now()
			logAttr := cacheInstance.LogAttr()
			timber.Info("setting up", logAttr)
			setup()
			timber.InfoSince(start, "setup", logAttr)
		})
	}
	wg.Wait()

	timber.InfoSince(start, "starting server")
	server := &http.Server{
		Addr:         ":8000",
		Handler:      middleware.Log(mux),
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		timber.Fatal(err, "failed to start router")
	}
}

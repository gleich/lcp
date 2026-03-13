package main

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/api/applemusic"
	"go.mattglei.ch/lcp/internal/api/github"
	"go.mattglei.ch/lcp/internal/api/steam"
	"go.mattglei.ch/lcp/internal/api/workouts"
	"go.mattglei.ch/lcp/internal/health"
	"go.mattglei.ch/lcp/internal/middleware"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/internal/tasks"
)

func main() {
	task, start := tasks.StartServer.Start()
	secrets.Load()

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
	mux.HandleFunc("/health", health.Endpoint)
	github.Setup(mux)
	workouts.Setup(mux, client, rdb)
	steam.Setup(mux, client, rdb)
	applemusic.Setup(mux, client, rdb)

	task.InfoSince("starting server", start)
	server := &http.Server{
		Addr:         ":8000",
		Handler:      middleware.LogRequest(mux),
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		task.ErrorSince(err, "failed to start server", start)
	}
}

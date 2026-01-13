package main

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"go.mattglei.ch/lcp/internal/apis/applemusic"
	"go.mattglei.ch/lcp/internal/apis/github"
	"go.mattglei.ch/lcp/internal/apis/steam"
	"go.mattglei.ch/lcp/internal/apis/workouts"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/timber"
)

func main() {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		timber.Fatal(err, "failed to load new york timezone")
	}
	timber.Timezone(ny)
	timber.TimeFormat("01/02 03:04:05 PM MST")

	timber.Info("booted")

	secrets.Load()

	var (
		client = http.Client{Timeout: 20 * time.Second}
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
	github.Setup(mux)
	workouts.Setup(mux, &client, rdb)
	steam.Setup(mux, &client, rdb)
	applemusic.Setup(mux, &client, rdb)

	timber.Info("starting server")
	server := &http.Server{
		Addr:         ":8000",
		Handler:      mux,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	err = server.ListenAndServe()
	if err != nil {
		timber.Fatal(err, "failed to start router")
	}
}

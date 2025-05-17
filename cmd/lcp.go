package main

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/apis/applemusic"
	"go.mattglei.ch/lcp/internal/apis/github"
	"go.mattglei.ch/lcp/internal/apis/steam"
	"go.mattglei.ch/lcp/internal/apis/workouts"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/timber"
)

var foo = "bar"

func main() {
	setupLogger()
	timber.Info("booted")

	secrets.Load()

	var (
		client = http.Client{}
		mux    = http.NewServeMux()
		rdb    = redis.NewClient(&redis.Options{
			Addr:     secrets.ENV.RedisAddress,
			Password: secrets.ENV.RedisPassword,
			DB:       0,
		})
	)

	mux.HandleFunc("/", rootRedirect)
	github.Setup(mux)
	workouts.Setup(mux, &client, rdb)
	steam.Setup(mux, &client, rdb)
	applemusic.Setup(mux, &client, rdb)

	timber.Info("starting server")
	err := http.ListenAndServe(":8000", mux)
	if err != nil {
		timber.Fatal(err, "failed to start router")
	}
}

func setupLogger() {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		timber.Fatal(err, "failed to load new york timezone")
	}
	timber.Timezone(ny)
	timber.TimeFormat("01/02 03:04:05 PM MST")
}

func rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://mattglei.ch/writing/lcp", http.StatusPermanentRedirect)
}

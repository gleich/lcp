package main

import (
	"net/http"
	"time"

	"go.mattglei.ch/lcp-2/internal/apis/applemusic"
	"go.mattglei.ch/lcp-2/internal/apis/github"
	"go.mattglei.ch/lcp-2/internal/apis/steam"
	"go.mattglei.ch/lcp-2/internal/apis/workouts"
	"go.mattglei.ch/lcp-2/internal/secrets"
	"go.mattglei.ch/timber"
)

func main() {
	setupLogger()
	timber.Info("booted")

	secrets.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootRedirect)

	github.Setup(mux)
	workouts.Setup(mux)
	steam.Setup(mux)
	applemusic.Setup(mux)

	timber.Info("starting server")
	err := http.ListenAndServe(":8000", mux)
	if err != nil {
		timber.Fatal(err, "failed to start router")
	}
}

func setupLogger() {
	nytime, err := time.LoadLocation("America/New_York")
	if err != nil {
		timber.Fatal(err, "failed to load new york timezone")
	}
	timber.Timezone(nytime)
	timber.TimeFormat("01/02 03:04:05 PM MST")
}

func rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://mattglei.ch/lcp", http.StatusPermanentRedirect)
}

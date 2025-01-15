package main

import (
	"net/http"
	"time"

	"pkg.mattglei.ch/lcp-2/internal/apis/applemusic"
	"pkg.mattglei.ch/lcp-2/internal/apis/github"
	"pkg.mattglei.ch/lcp-2/internal/apis/steam"
	"pkg.mattglei.ch/lcp-2/internal/apis/strava"
	"pkg.mattglei.ch/lcp-2/internal/secrets"
	"pkg.mattglei.ch/timber"
)

func main() {
	setupLogger()
	timber.Info("booted")

	secrets.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootRedirect)

	github.Setup(mux)
	strava.Setup(mux)
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
	timber.SetTimezone(nytime)
	timber.SetTimeFormat("01/02 03:04:05 PM MST")
}

func rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://mattglei.ch/lcp", http.StatusPermanentRedirect)
}

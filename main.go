package main

import (
	"net/http"

	"github.com/caarlos0/env/v11"
	"github.com/gleich/lcp/pkg/cache"
	"github.com/gleich/lcp/pkg/secrets"
	"github.com/gleich/lcp/pkg/strava"
	"github.com/gleich/lumber/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	lumber.Info("booted")

	err := godotenv.Load()
	if err != nil {
		lumber.Fatal(err, "Error loading .env file")
	}
	loadedSecrets, err := env.ParseAs[secrets.Secrets]()
	if err != nil {
		lumber.Fatal(err, "parsing required env vars failed")
	}

	activities, err := strava.FetchActivities(loadedSecrets)
	if err != nil {
		lumber.Fatal(err, "failed to do initial fetch on strava activities")
	}
	stravaCache := cache.New("strava", activities)
	lumber.Success("init strava cache")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.HandleFunc("/", rootRedirect)
	r.Get("/strava/cache", cache.Route(&stravaCache, loadedSecrets))
	err = http.ListenAndServe(":8000", r)
	if err != nil {
		lumber.Fatal(err, "failed to start router")
	}
}

func rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://mattglei.ch", http.StatusTemporaryRedirect)
}

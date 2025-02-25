package hevy

import (
	"net/http"
	"time"

	"go.mattglei.ch/lcp-2/internal/cache"
	"go.mattglei.ch/timber"
)

const LOG_PREFIX = "[hevy]"

func Setup(mux *http.ServeMux) {
	client := http.Client{}
	workouts, err := fetchWorkouts(&client)
	if err != nil {
		timber.Error(err, "initial fetch of hevy workouts failed")
	}

	hevyCache := cache.New("hevy", workouts, err == nil)
	mux.HandleFunc("GET /hevy", hevyCache.ServeHTTP)
	go cache.UpdatePeriodically(hevyCache, &client, fetchWorkouts, 5*time.Minute)
	timber.Done(LOG_PREFIX, "setup cache and endpoint")
}

package steam

import (
	"net/http"
	"time"

	"pkg.mattglei.ch/lcp-2/internal/cache"
	"pkg.mattglei.ch/timber"
)

func Setup(mux *http.ServeMux) {
	client := http.Client{}
	games, err := fetchRecentlyPlayedGames(&client)
	if err != nil {
		timber.Error(err, "initial fetch of games failed")
	}

	steamCache := cache.New("steam", games, err == nil)
	mux.HandleFunc("GET /steam", steamCache.ServeHTTP)
	go cache.UpdatePeriodically(steamCache, &client, fetchRecentlyPlayedGames, 5*time.Minute)
	timber.Done("setup steam cache")
}

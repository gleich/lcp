package steam

import (
	"net/http"
	"time"

	"pkg.mattglei.ch/lcp-2/internal/cache"
	"pkg.mattglei.ch/timber"
)

func Setup(mux *http.ServeMux) {
	games, err := fetchRecentlyPlayedGames()
	if err != nil {
		timber.Error(err, "initial fetch of games failed")
	}

	steamCache := cache.New("steam", games, err == nil)
	mux.HandleFunc("GET /steam", steamCache.ServeHTTP)
	go steamCache.UpdatePeriodically(fetchRecentlyPlayedGames, 5*time.Minute)
	timber.Done("setup steam cache")
}

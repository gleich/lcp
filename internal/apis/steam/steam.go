package steam

import (
	"net/http"
	"time"

	"go.mattglei.ch/lcp-2/internal/cache"
	"go.mattglei.ch/timber"
)

const LOG_PREFIX = "[steam]"

func Setup(mux *http.ServeMux) {
	client := http.Client{}
	games, err := fetchRecentlyPlayedGames(&client)
	if err != nil {
		timber.Error(err, "initial fetch of games failed")
	}

	steamCache := cache.New("steam", games, err == nil)
	mux.HandleFunc("GET /steam", steamCache.ServeHTTP)
	go cache.UpdatePeriodically(steamCache, &client, fetchRecentlyPlayedGames, 5*time.Minute)
	timber.Done(LOG_PREFIX, "setup cache and endpoint")
}

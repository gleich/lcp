package github

import (
	"context"
	"net/http"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"pkg.mattglei.ch/lcp-2/internal/cache"
	"pkg.mattglei.ch/lcp-2/internal/secrets"
	"pkg.mattglei.ch/timber"
)

func Setup(mux *http.ServeMux) {
	githubTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: secrets.ENV.GitHubAccessToken},
	)
	githubHttpClient := oauth2.NewClient(context.Background(), githubTokenSource)
	githubClient := githubv4.NewClient(githubHttpClient)

	pinnedRepos, err := fetchPinnedRepos(githubClient)
	if err != nil {
		timber.Error(err, "fetching initial pinned repos failed")
	}

	githubCache := cache.New("github", pinnedRepos, err == nil)
	mux.HandleFunc("GET /github", githubCache.ServeHTTP)
	go cache.UpdatePeriodically(githubCache, githubClient, fetchPinnedRepos, 1*time.Minute)
	timber.Done("setup github cache")
}

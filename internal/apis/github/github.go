package github

import (
	"context"
	"net/http"
	"time"

	"github.com/shurcooL/githubv4"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/timber"
	"golang.org/x/oauth2"
)

const cacheInstance = cache.GitHub

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

	githubCache := cache.New(cacheInstance, pinnedRepos, err == nil)
	mux.HandleFunc("GET /github", githubCache.ServeHTTP)
	go cache.UpdatePeriodically(githubCache, githubClient, fetchPinnedRepos, 30*time.Second)

	timber.Done(cacheInstance.LogPrefix(), "setup cache and endpoint")
}

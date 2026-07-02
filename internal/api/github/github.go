package github

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/shurcooL/githubv4"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/secrets"
	"golang.org/x/oauth2"
)

const cacheInstance = cache.GitHub

var logger = sync.OnceValue(func() *zerolog.Logger { return cacheInstance.Logger() })

func Setup(mux *http.ServeMux) {
	githubTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: secrets.ENV.GitHubAccessToken},
	)
	githubHttpClient := oauth2.NewClient(context.Background(), githubTokenSource)
	githubClient := githubv4.NewClient(githubHttpClient)

	pinnedRepos, err := fetchPinnedRepos(githubClient)
	if err != nil {
		logger().Error().Err(err).Msg("fetching initial pinned repos failed")
	}

	githubCache := cache.New(cacheInstance, pinnedRepos, err == nil)
	githubCache.Endpoints(mux)
	go cache.UpdatePeriodically(githubCache, githubClient, fetchPinnedRepos, 5*time.Second)
}

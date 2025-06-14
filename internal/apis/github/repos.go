package github

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/shurcooL/githubv4"
	"go.mattglei.ch/lcp/internal/apis"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

type pinnedItemsQuery struct {
	Viewer struct {
		PinnedItems struct {
			Nodes []struct {
				Repository struct {
					Name  githubv4.String
					Owner struct {
						Login githubv4.String
					}
					PrimaryLanguage struct {
						Name  githubv4.String
						Color githubv4.String
					}
					Description githubv4.String
					UpdatedAt   githubv4.DateTime
					IsPrivate   githubv4.Boolean
					ID          githubv4.ID
					URL         githubv4.URI
				} `graphql:"... on Repository"`
			}
		} `graphql:"pinnedItems(first: 6, types: REPOSITORY)"`
	}
}

func fetchPinnedRepos(client *githubv4.Client) ([]lcp.GitHubRepository, error) {
	var query pinnedItemsQuery
	err := client.Query(context.Background(), &query, nil)
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		timber.Warning(cacheInstance.LogPrefix(), "connection timed out for getting pinned repos")
		return []lcp.GitHubRepository{}, apis.ErrWarning
	}
	if err != nil && (errors.Is(err, syscall.ECONNRESET) ||
		strings.Contains(err.Error(), "connection reset by peer")) {
		timber.Warning(cacheInstance.LogPrefix(),
			"connection reset by peer while getting pinned repos")
		return []lcp.GitHubRepository{}, apis.ErrWarning
	}
	if err != nil {
		return nil, fmt.Errorf("%w querying github's graphql API failed", err)
	}

	var repositories []lcp.GitHubRepository
	for _, node := range query.Viewer.PinnedItems.Nodes {
		repositories = append(repositories, lcp.GitHubRepository{
			Name:          string(node.Repository.Name),
			Owner:         string(node.Repository.Owner.Login),
			Language:      string(node.Repository.PrimaryLanguage.Name),
			LanguageColor: string(node.Repository.PrimaryLanguage.Color),
			Description:   string(node.Repository.Description),
			UpdatedAt:     node.Repository.UpdatedAt.Time,
			ID:            fmt.Sprint(node.Repository.ID),
			URL:           fmt.Sprint(node.Repository.URL.URL),
		})
	}
	return repositories, nil
}

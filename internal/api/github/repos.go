package github

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/shurcooL/githubv4"
	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/tasks"
	"go.mattglei.ch/lcp/pkg/lcp"
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
	task := tasks.Cache.GitHub.FetchPinnedRepos
	var query pinnedItemsQuery
	err := client.Query(context.Background(), &query, nil)
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		task.Warn("connection timed out for getting pinned repos")
		return []lcp.GitHubRepository{}, api.ErrWarning
	}
	if err != nil && (errors.Is(err, syscall.ECONNRESET) ||
		strings.Contains(err.Error(), "connection reset by peer")) {
		task.Warn("connection reset")
		return []lcp.GitHubRepository{}, api.ErrWarning
	}
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "non-200 OK status code") {
			if strings.Contains(errMsg, "500 Internal Server Error body") {
				task.Warn("500 Interval Server Error body")
				return []lcp.GitHubRepository{}, api.ErrWarning
			}
			if strings.Contains(errMsg, "502 Bad Gateway body") {
				task.Warn("502 Bad Gateway body")
				return []lcp.GitHubRepository{}, api.ErrWarning
			}
			if strings.Contains(errMsg, "503 Service Unavailable body") {
				task.Warn("503 service unavailable")
				return []lcp.GitHubRepository{}, api.ErrWarning
			}
		}
		return nil, fmt.Errorf("querying github's graphql API: %w", err)
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
			URL:           node.Repository.URL.String(),
		})
	}
	return repositories, nil
}

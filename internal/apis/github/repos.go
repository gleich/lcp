package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"pkg.mattglei.ch/lcp-2/pkg/models"
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

func fetchPinnedRepos(client *githubv4.Client) ([]models.GitHubRepository, error) {
	var query pinnedItemsQuery
	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		return nil, fmt.Errorf("%v querying github's graphql API failed", err)
	}

	var repositories []models.GitHubRepository
	for _, node := range query.Viewer.PinnedItems.Nodes {
		repositories = append(repositories, models.GitHubRepository{
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

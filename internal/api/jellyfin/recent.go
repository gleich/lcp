package jellyfin

import (
	"fmt"
	"net/http"
	"net/url"

	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/pkg/lcp"
)

type itemsResponse struct {
	Items []struct {
		ID                string `json:"Id"`
		Name              string `json:"Name"`
		IndexNumber       int    `json:"IndexNumber"`
		ParentIndexNumber int    `json:"ParentIndexNumber"`
		Type              string `json:"Type"`
		SeriesName        string `json:"SeriesName"`
		SeriesID          string `json:"SeriesId"`
		ImageTags         struct {
			Primary string `json:"Primary"`
		} `json:"ImageTags"`
		ImageBlurHashes struct {
			Primary map[string]string `json:"Primary"`
		} `json:"ImageBlurHashes"`
	} `json:"Items"`
}

func fetchRecentlyPlayed(client *http.Client) ([]lcp.JellyfinItem, error) {
	params := url.Values{
		"userId":           {secrets.ENV.JellyfinUserID},
		"IncludeItemTypes": {"Movie,Episode"},
		"SortBy":           {"DatePlayed"},
		"SortOrder":        {"Descending"},
		"Recursive":        {"true"},
		"Limit":            {"10"},
		"Fields":           {"ProviderIds"},
		"EnableImageTypes": {"Primary"},
	}
	response, err := sendJellyfinAPIRequest[itemsResponse](client, "/Items?"+params.Encode())
	if err != nil {
		return []lcp.JellyfinItem{}, fmt.Errorf("fetching recently played from jellyfin: %w", err)
	}

	items := make([]lcp.JellyfinItem, 0, len(response.Items))
	for _, item := range response.Items {
		normalized := lcp.JellyfinItem{
			ID:   item.ID,
			Type: item.Type,
			Name: item.Name,
		}

		if item.Type == "Episode" {
			normalized.SeriesName = item.SeriesName
			normalized.SeriesID = item.SeriesID
			normalized.SeasonNumber = item.ParentIndexNumber
			normalized.EpisodeNumber = item.IndexNumber
		}

		if item.ImageTags.Primary != "" {
			imageID := item.ID
			if item.Type == "Episode" && item.SeriesID != "" {
				imageID = item.SeriesID
			}
			imageURL := fmt.Sprintf("https://gleich.tv/Items/%s/Images/Primary", imageID)
			normalized.ImageURL = &imageURL

			if imageID == item.ID {
				if blurhash, ok := item.ImageBlurHashes.Primary[item.ImageTags.Primary]; ok {
					normalized.ImageBlurhash = &blurhash
				}
			}
		}

		items = append(items, normalized)
	}

	return items, nil
}

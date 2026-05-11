package jellyfin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"

	"github.com/minio/minio-go/v7"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/pkg/lcp"
)

const bucketName = "jellyfin-images"

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

func fetchRecentlyPlayed(
	client *http.Client,
	minioClient *minio.Client,
) ([]lcp.JellyfinItem, error) {
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

	var imageIDs []string
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

			err := uploadImage(client, minioClient, imageID)
			if err != nil {
				return []lcp.JellyfinItem{}, fmt.Errorf("uploading image for %s: %w", imageID, err)
			}
			imageIDs = append(imageIDs, imageID)

			imageURL := fmt.Sprintf(
				"https://%s/%s/%s.jpg",
				secrets.ENV.MinioEndpoint,
				bucketName,
				imageID,
			)
			normalized.ImageURL = &imageURL

			if imageID == item.ID {
				if blurhash, ok := item.ImageBlurHashes.Primary[item.ImageTags.Primary]; ok {
					normalized.ImageBlurhash = &blurhash
				}
			}
		}

		items = append(items, normalized)
	}

	err = removeOldImages(minioClient, imageIDs)
	if err != nil {
		return []lcp.JellyfinItem{}, fmt.Errorf("removing old jellyfin images: %w", err)
	}

	return items, nil
}

func uploadImage(client *http.Client, minioClient *minio.Client, imageID string) error {
	objectKey := imageID + ".jpg"

	_, err := minioClient.StatObject(
		context.Background(),
		bucketName,
		objectKey,
		minio.StatObjectOptions{},
	)
	if err == nil {
		return nil
	}

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://gleich.tv/Items/%s/Images/Primary?format=jpg", imageID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Emby-Token", secrets.ENV.JellyfinKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading image data: %w", err)
	}

	_, err = minioClient.PutObject(
		context.Background(),
		bucketName,
		objectKey,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: "image/jpeg"},
	)
	if err != nil {
		return fmt.Errorf("uploading to minio: %w", err)
	}
	return nil
}

func removeOldImages(minioClient *minio.Client, imageIDs []string) error {
	validKeys := make([]string, len(imageIDs))
	for i, id := range imageIDs {
		validKeys[i] = id + ".jpg"
	}

	objects := minioClient.ListObjects(
		context.Background(),
		bucketName,
		minio.ListObjectsOptions{},
	)
	for object := range objects {
		if object.Err != nil {
			return fmt.Errorf("listing minio objects: %w", object.Err)
		}
		if !slices.Contains(validKeys, object.Key) {
			err := minioClient.RemoveObject(
				context.Background(),
				bucketName,
				object.Key,
				minio.RemoveObjectOptions{},
			)
			if err != nil {
				return fmt.Errorf("removing minio object: %w", err)
			}
		}
	}
	return nil
}

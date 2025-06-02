package strava

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/minio/minio-go/v7"
	"go.mattglei.ch/lcp/internal/apis"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

const bucketName = "mapbox-maps"

func FetchMap(client *http.Client, polyline string) ([]byte, error) {
	var (
		lineWidth = 2.0
		lineColor = "000"
		width     = 462
		height    = 252
		style     = "mattgleich/clxxsfdfm002401qj7jcxh47e"
		params    = url.Values{"access_token": {secrets.ENV.MapboxAccessToken}}
		url       = fmt.Sprintf(
			"https://api.mapbox.com/styles/v1/%s/static/path-%f+%s(%s)/auto/%dx%d@2x?%s",
			style,
			lineWidth,
			lineColor,
			url.QueryEscape(polyline),
			width,
			height,
			params.Encode(),
		)
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w failed to create get request to %s", err, url)
	}

	b, err := apis.Request(logPrefix, client, req)
	if err != nil {
		timber.Error(err, "failed to send request to", url)
		return nil, fmt.Errorf("%w failed to send request to %s", err, url)
	}

	return b, nil
}

func UploadMap(minioClient *minio.Client, id string, data []byte) error {
	var (
		reader = bytes.NewReader(data)
		size   = int64(len(data))
	)

	_, err := minioClient.PutObject(
		context.Background(),
		bucketName,
		fmt.Sprintf("%s.png", id),
		reader,
		size,
		minio.PutObjectOptions{ContentType: "image/png"},
	)
	if err != nil {
		return fmt.Errorf("%w failed to upload to minio", err)
	}
	return nil
}

func RemoveOldMaps(minioClient *minio.Client, activities []lcp.Workout) error {
	var validKeys []string
	for _, activity := range activities {
		validKeys = append(validKeys, fmt.Sprintf("%s.png", activity.ID))
	}

	objects := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{})
	for object := range objects {
		if object.Err != nil {
			return fmt.Errorf("%w failed to load object", object.Err)
		}
		if !slices.Contains(validKeys, object.Key) {
			err := minioClient.RemoveObject(
				context.Background(),
				bucketName,
				object.Key,
				minio.RemoveObjectOptions{},
			)
			if err != nil {
				return fmt.Errorf("%w failed to remove object", err)
			}
		}
	}
	return nil
}

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
)

const bucketName = "mapbox-maps"

func FetchMap(client *http.Client, polyline string) ([]byte, error) {
	var (
		lineWidth = 2.0
		lineColor = "000"
		width     = 462
		height    = 252
		style     = "mattgleich/clxxsfdfm002401qj7jcxh47e"
		params    = url.Values{
			"access_token": {secrets.ENV.MapboxAccessToken},
			"padding":      {"28"},
			"attribution":  {"false"},
			"logo":         {"false"},
		}
		url = fmt.Sprintf(
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
		return nil, fmt.Errorf("creating request (url: %s): %w", url, err)
	}

	b, err := apis.Request(logPrefix, client, req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
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
		return fmt.Errorf("uploading to minio: %w", err)
	}
	return nil
}

func RemoveOldMaps(minioClient *minio.Client, workouts []lcp.Workout) error {
	var validKeys []string
	for _, activity := range workouts {
		validKeys = append(validKeys, fmt.Sprintf("%s.png", activity.ID))
	}

	objects := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{})
	for object := range objects {
		if object.Err != nil {
			return fmt.Errorf("loading minio objects: %w", object.Err)
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

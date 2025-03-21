package strava

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/minio/minio-go/v7"
	"go.mattglei.ch/lcp-2/internal/apis"
	"go.mattglei.ch/lcp-2/internal/secrets"
	"go.mattglei.ch/lcp-2/pkg/lcp"
	"go.mattglei.ch/timber"
)

const bucketName = "mapbox-maps"

func fetchMap(polyline string, client *http.Client) []byte {
	var (
		lineWidth = 2.0
		lineColor = "000"
		width     = 462
		height    = 252
		params    = url.Values{"access_token": {secrets.ENV.MapboxAccessToken}}
		url       = fmt.Sprintf(
			"https://api.mapbox.com/styles/v1/mattgleich/clxxsfdfm002401qj7jcxh47e/static/path-%f+%s(%s)/auto/%dx%d@2x?%s",
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
		timber.Error(err, "failed to create get request to", url)
	}

	b, err := apis.Request(logPrefix, client, req)
	if err != nil {
		timber.Error(err, "failed to send request to", url)
	}

	return b
}

func uploadMap(minioClient minio.Client, id uint64, data []byte) {
	var (
		reader = bytes.NewReader(data)
		size   = int64(len(data))
	)

	_, err := minioClient.PutObject(
		context.Background(),
		bucketName,
		fmt.Sprintf("%d.png", id),
		reader,
		size,
		minio.PutObjectOptions{ContentType: "image/png"},
	)
	if err != nil {
		timber.Error(err, "failed to upload to minio")
	}
}

func removeOldMaps(minioClient minio.Client, activities []lcp.Workout) {
	var validKeys []string
	for _, activity := range activities {
		validKeys = append(validKeys, fmt.Sprintf("%s.png", activity.ID))
	}

	objects := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{})
	for object := range objects {
		if object.Err != nil {
			timber.Error(object.Err, "failed to load object")
			return
		}
		if !slices.Contains(validKeys, object.Key) {
			err := minioClient.RemoveObject(
				context.Background(),
				bucketName,
				object.Key,
				minio.RemoveObjectOptions{},
			)
			if err != nil {
				timber.Error(err, "failed to remove object")
				return
			}
		}
	}
}

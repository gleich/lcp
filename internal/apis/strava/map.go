package strava

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/minio/minio-go/v7"
	"pkg.mattglei.ch/lcp-2/internal/secrets"
	"pkg.mattglei.ch/timber"
)

const bucketName = "mapbox-maps"

func fetchMap(polyline string) []byte {
	var (
		lineWidth = 2.0
		lineColor = "000"
		width     = 440
		height    = 240
		params    = url.Values{"access_token": {secrets.ENV.MapboxAccessToken}}
	)
	url := fmt.Sprintf(
		"https://api.mapbox.com/styles/v1/mattgleich/clxxsfdfm002401qj7jcxh47e/static/path-%f+%s(%s)/auto/%dx%d@2x?%s",
		lineWidth,
		lineColor,
		url.QueryEscape(polyline),
		width,
		height,
		params.Encode(),
	)
	resp, err := http.Get(url)
	if err != nil {
		timber.Error(err, "failed to fetch mapbox image with polyline", url)
		return nil
	}

	var b bytes.Buffer
	_, err = b.ReadFrom(resp.Body)
	if err != nil {
		timber.Error(err, "failed to read data from request")
		return nil
	}

	return b.Bytes()
}

func uploadMap(minioClient minio.Client, id uint64, data []byte) {
	reader := bytes.NewReader(data)
	size := int64(len(data))

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

func removeOldMaps(minioClient minio.Client, activities []activity) {
	var validKeys []string
	for _, activity := range activities {
		validKeys = append(validKeys, fmt.Sprintf("%d.png", activity.ID))
	}

	objects := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{})
	for object := range objects {
		if object.Err != nil {
			timber.Error(object.Err, "failed to load object")
			return
		}
		var validObject bool
		for _, validKey := range validKeys {
			if validKey == object.Key {
				validObject = true
				break
			}
		}
		if !validObject {
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

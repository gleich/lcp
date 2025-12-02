package workouts

import (
	"fmt"
	"image/png"
	"net/http"
	"sort"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/apis/workouts/hevy"
	"go.mattglei.ch/lcp/internal/apis/workouts/strava"
	"go.mattglei.ch/lcp/internal/images"
	"go.mattglei.ch/lcp/pkg/lcp"
)

func fetch(
	client *http.Client,
	minioClient *minio.Client,
	rdb *redis.Client,
	stravaTokens strava.Tokens,
) ([]lcp.Workout, error) {
	stravaActivities, err := strava.FetchActivities(client, minioClient, rdb, stravaTokens)
	if err != nil {
		return []lcp.Workout{}, err
	}

	hevyWorkouts, err := hevy.FetchWorkouts(client)
	if err != nil {
		return []lcp.Workout{}, err
	}

	activities := []lcp.Workout{}
	activities = append(activities, hevyWorkouts...)

	for _, s := range stravaActivities {
		conflict := false
		for _, h := range hevyWorkouts {
			diff := s.StartDate.Sub(h.StartDate)
			if diff < 0 {
				diff = -diff
			}
			if diff < time.Minute {
				conflict = true
				break
			}
		}
		if !conflict {
			activities = append(activities, s)
		}
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].StartDate.After(activities[j].StartDate)
	})

	// only store the first 20 activities
	activities = activities[:20]

	// fill in data for collected strava activities. this is done to keep the number of API requests
	// to strava to a minimum. Rate limits were getting hit when making requests for all strava
	// activities, so this should help mitigate that (especially when having to restart the
	// application during updates).
	for i := range activities {
		activity := &activities[i]
		if activity.Platform != "strava" {
			continue
		}

		details, err := strava.FetchActivityDetails(client, activity.ID, stravaTokens)
		if err != nil {
			return nil, fmt.Errorf(
				"%w failed to fetch activity details for activity with ID of %s",
				err,
				activity.ID,
			)
		}
		activity.Calories = details.Calories

		heartrateStream, err := strava.FetchHeartrate(client, activity.ID, stravaTokens)
		if err != nil {
			return nil, fmt.Errorf(
				"%w failed to fetch HR data for activity with ID of %s",
				err,
				activity.ID,
			)
		}
		activity.HeartrateData = heartrateStream

		if activity.HasMap {
			mapData, err := strava.FetchMap(client, activity.MapPolyline)
			if err != nil {
				return nil, fmt.Errorf("%w failed to fetch map", err)
			}
			err = strava.UploadMap(minioClient, activity.ID, mapData)
			if err != nil {
				return nil, fmt.Errorf("%w failed to upload map", err)
			}
			imgURL := fmt.Sprintf(
				"https://s3.mattglei.ch/mapbox-maps/%s.png",
				activity.ID,
			)
			mapBlurHash, err := images.BlurHash(client, rdb, imgURL, png.Decode)
			if err != nil {
				return nil, fmt.Errorf("%w failed to create blur hash for image", err)
			}
			activity.MapBlurImage = &mapBlurHash
			activity.MapImageURL = &imgURL

			location, err := strava.FetchLocation(client, *activity)
			if err != nil {
				return nil, fmt.Errorf(
					"%w failed to fetch location data for %s",
					err,
					activity.Name,
				)
			}
			activity.Location = location
		}
	}

	err = strava.RemoveOldMaps(minioClient, activities)
	if err != nil {
		return nil, fmt.Errorf("%w failed to remove old maps", err)
	}

	return activities, nil
}

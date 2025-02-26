package activities

import (
	"net/http"

	"github.com/minio/minio-go/v7"
	"go.mattglei.ch/lcp-2/internal/apis/activities/hevy"
	"go.mattglei.ch/lcp-2/internal/apis/activities/strava"
	"go.mattglei.ch/lcp-2/pkg/lcp"
	"go.mattglei.ch/timber"
)

func fetch(
	client *http.Client,
	minioClient minio.Client,
	stravaTokens strava.Tokens,
) ([]lcp.Activity, error) {
	stravaActivities, err := strava.FetchActivities(client, minioClient, stravaTokens)
	if err != nil {
		return []lcp.Activity{}, err
	}

	hevyWorkouts, err := hevy.FetchWorkouts(client)
	if err != nil {
		return []lcp.Activity{}, err
	}

	for _, h := range hevyWorkouts {
		timber.Debug(h.Name)
	}

	activities := []lcp.Activity{}
	activities = append(activities, hevyWorkouts...)

	for _, s := range stravaActivities {
		conflict := false
		for _, h := range hevyWorkouts {
			if s.StartDate.Equal(h.StartDate) {
				conflict = true
				break
			}
		}
		if !conflict {
			activities = append(activities, s)
		}
	}

	return activities, nil
}

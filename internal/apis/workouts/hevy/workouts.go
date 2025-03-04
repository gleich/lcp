package hevy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mattglei.ch/lcp-2/internal/secrets"
	"go.mattglei.ch/lcp-2/pkg/lcp"
	"go.mattglei.ch/timber"
)

type workoutsResponse struct {
	Workouts []struct {
		ID        string             `json:"id"`
		Title     string             `json:"title"`
		StartTime time.Time          `json:"start_time"`
		EndTime   time.Time          `json:"end_time"`
		CreatedAt time.Time          `json:"created_at"`
		Exercises []lcp.HevyExercise `json:"exercises"`
	} `json:"workouts"`
}

func FetchWorkouts(client *http.Client) ([]lcp.Workout, error) {
	params := url.Values{"api-key": {secrets.ENV.HevyAccessToken}}
	workouts, err := sendHevyAPIRequest[workoutsResponse](
		client,
		fmt.Sprintf("/v1/workouts?%s", params.Encode()),
	)
	if err != nil {
		return []lcp.Workout{}, fmt.Errorf("%w ", err)
	}

	var activities []lcp.Workout
	for _, workout := range workouts.Workouts {
		totalVolume := 0.0
		sets := 0
		for _, exercise := range workout.Exercises {
			multiplier := 1.0
			title := strings.ToLower(exercise.Title)
			if strings.Contains(title, "(dumbbell)") {
				multiplier = 2.0
			}
			for _, set := range exercise.Sets {
				totalVolume += multiplier * set.WeightKg * float64(set.Reps)
				sets++
			}
		}
		timber.Debug(totalVolume)
		activities = append(activities, lcp.Workout{
			Platform:      "hevy",
			Name:          workout.Title,
			StartDate:     workout.StartTime,
			MovingTime:    uint32(workout.EndTime.Sub(workout.StartTime).Seconds()),
			SportType:     "WeightTraining",
			Timezone:      "(GMT-05:00) America/New_York",
			HasMap:        false,
			ID:            workout.ID,
			HasHeartrate:  false,
			HevyExercises: workout.Exercises,
			HevyVolumeKG:  totalVolume,
			HevySetCount:  sets,
		})
	}

	return activities, nil
}

package hevy

import (
	"fmt"
	"net/http"
	"net/url"

	"go.mattglei.ch/lcp-2/internal/secrets"
	"go.mattglei.ch/lcp-2/pkg/lcp"
)

type workoutsResponse struct {
	Workouts []lcp.HevyWorkout `json:"workouts"`
}

func fetchWorkouts(client *http.Client) ([]lcp.HevyWorkout, error) {
	params := url.Values{"api-key": {secrets.ENV.HevyAccessToken}}
	workouts, err := sendHevyAPIRequest[workoutsResponse](
		client,
		fmt.Sprintf("/v1/workouts?%s", params.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("%w ", err)
	}

	return workouts.Workouts, nil
}

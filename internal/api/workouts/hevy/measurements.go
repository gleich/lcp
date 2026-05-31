package hevy

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type bodyMeasurementsResponse struct {
	Page             int `json:"page"`
	PageCount        int `json:"page_count"`
	BodyMeasurements []struct {
		ID        int       `json:"id"`
		Date      string    `json:"date"`
		WeightKg  float64   `json:"weight_kg"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"body_measurements"`
}

// returns the body weight entry in kg
func fetchBodyWeight(client *http.Client) (float64, error) {
	const endpoint = "/v1/body_measurements?"

	// initial fetch to get page count
	params := url.Values{"pageSize": {"1"}}
	measurements, err := sendHevyRequest[bodyMeasurementsResponse](client, endpoint+params.Encode())
	if err != nil {
		return 0.0, fmt.Errorf("fetching initial body measurements: %w", err)
	}

	// actually fetch the body weight entry
	params.Add("page", strconv.Itoa(int(measurements.PageCount)))
	measurements, err = sendHevyRequest[bodyMeasurementsResponse](client, endpoint+params.Encode())
	if err != nil {
		return 0.0, fmt.Errorf("fetching last page of initial body measurements: %w", err)
	}
	count := len(measurements.BodyMeasurements)
	if count != 1 {
		return 0.0, fmt.Errorf(
			"incorrect number of body measurements returned (%d returned): %w",
			count,
			err,
		)
	}
	return measurements.BodyMeasurements[0].WeightKg, nil
}

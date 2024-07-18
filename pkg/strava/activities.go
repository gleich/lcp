package strava

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gleich/lumber/v2"
)

type Activity struct {
	Name      string    `json:"name"`
	SportType string    `json:"sport_type"`
	StartDate time.Time `json:"start_date"`
	Timezone  string    `json:"timezone"`
	Map       struct {
		SummaryPolyline string `json:"summary_polyline"`
	} `json:"map"`
	Trainer            bool    `json:"trainer"`
	Commute            bool    `json:"commute"`
	Private            bool    `json:"private"`
	AverageSpeed       float32 `json:"average_speed"`
	MaxSpeed           float32 `json:"max_speed"`
	AverageTemp        int32   `json:"average_temp,omitempty"`
	AverageCadence     float32 `json:"average_cadence,omitempty"`
	AverageWatts       float32 `json:"average_watts,omitempty"`
	DeviceWatts        bool    `json:"device_watts,omitempty"`
	AverageHeartrate   float32 `json:"average_heartrate,omitempty"`
	TotalElevationGain float32 `json:"total_elevation_gain"`
	MovingTime         uint32  `json:"moving_time"`
	SufferScore        float32 `json:"suffer_score,omitempty"`
	PrCount            uint32  `json:"pr_count"`
	Distance           float32 `json:"distance"`
	ID                 uint64  `json:"id"`
}

func FetchActivities(tokens Tokens) []Activity {
	req, err := http.NewRequest("GET", "https://www.strava.com/api/v3/athlete/activities", nil)
	if err != nil {
		lumber.Error(err, "Failed to create new request")
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+tokens.Access)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		lumber.Error(err, "Failed to send request for Strava activities")
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		lumber.Error(err, "reading response body failed")
		return nil
	}

	var activities []Activity
	err = json.Unmarshal(body, &activities)
	if err != nil {
		lumber.Error(err, "failed to parse json")
		lumber.Debug(string(body))
		return nil
	}

	return activities
}

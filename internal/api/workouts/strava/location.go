package strava

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/pkg/lcp"
)

type locationResponse struct {
	Results []struct {
		Components struct {
			Borough      string `json:"borough"`
			City         string `json:"city"`
			State        string `json:"state"`
			StateCode    string `json:"state_code"`
			Municipality string `json:"municipality"`
			Town         string `json:"town"`
			Village      string `json:"village"`
			Country      string `json:"country"`
		} `json:"components"`
	} `json:"results"`
}

func FetchLocation(client *http.Client, workout lcp.Workout) (*string, error) {
	latitude := workout.Latitude
	longitude := workout.Longitude
	if (latitude == 0 && longitude == 0) || !workout.HasMap {
		return nil, nil
	}

	params := url.Values{
		"key": {secrets.ENV.OpenCageDataKey},
		"q":   {fmt.Sprintf("%f,%f", latitude, longitude)},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.opencagedata.com/geocode/v1/json?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request for location: %w", err)
	}

	resp, err := api.RequestJSON[locationResponse](client, req, logger())
	if err != nil {
		return nil, fmt.Errorf("sending request for location: %w", err)
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no location results returned for %s", workout.Name)
	}
	components := resp.Results[0].Components

	componentsToTrim := []*string{
		&components.City,
		&components.Municipality,
		&components.Town,
		&components.Village,
	}
	prefixes := []string{"City of", "Town of", "Village of"}
	for _, component := range componentsToTrim {
		for _, prefix := range prefixes {
			*component = strings.TrimPrefix(*component, prefix)
		}
	}

	var locality string
	if components.Borough != "" {
		locality = fmt.Sprintf("%s, %s", components.Borough, components.City)
	} else if components.Town != "" {
		locality = components.Town
	} else if components.Municipality != "" {
		locality = components.Municipality
	} else if components.City != "" {
		locality = components.City
	} else if components.Village != "" {
		locality = components.Village
	} else {
		logger().Warn().
			Str("workout-name", workout.Name).
			Float32("latitude", latitude).
			Float32("longitude", longitude).
			Msg("unable to create location")
		return nil, nil
	}

	region := components.StateCode
	if components.Country != "United States of America" {
		region = components.Country
		if components.State != "" && components.State != locality {
			region = fmt.Sprintf("%s, %s", components.State, components.Country)
		}
	}

	location := strings.TrimSpace(fmt.Sprintf("%s, %s", locality, region))

	return &location, nil
}

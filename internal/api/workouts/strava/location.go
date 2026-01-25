package strava

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
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
		"q": {
			fmt.Sprintf("%f,%f", latitude, longitude),
		},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.opencagedata.com/geocode/v1/json?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request for location: %w", err)
	}

	resp, err := api.RequestJSON[locationResponse](logPrefix, client, req)
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

	var location string
	if components.Borough != "" {
		location = fmt.Sprintf(
			"%s, %s %s",
			components.Borough,
			components.City,
			components.StateCode,
		)
	} else if components.Town != "" {
		location = fmt.Sprintf("%s, %s", components.Town, components.State)
	} else if components.Municipality != "" {
		location = fmt.Sprintf("%s, %s", components.Municipality, components.State)
	} else if components.City != "" {
		location = fmt.Sprintf("%s, %s", components.City, components.State)
	} else if components.Village != "" {
		location = fmt.Sprintf("%s, %s", components.Village, components.State)
	} else {
		timber.Warningf("unable to create location for %s (%f %f)", workout.Name, latitude, longitude)
		return nil, nil
	}

	location = strings.TrimSpace(location)

	return &location, nil
}

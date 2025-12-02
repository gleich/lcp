package strava

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.mattglei.ch/lcp/internal/apis"
	"go.mattglei.ch/lcp/internal/secrets"
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

func fetchLocation(client *http.Client, stravaActivity activity) (*string, error) {
	latitude := stravaActivity.StartLatLong[0]
	longitude := stravaActivity.StartLatLong[1]
	if (latitude == 0 && longitude == 0) ||
		stravaActivity.Map.SummaryPolyline == "" {
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
		return nil, fmt.Errorf("%w failed to create request for location", err)
	}

	resp, err := apis.RequestJSON[locationResponse](logPrefix, client, req)
	if err != nil {
		return nil, fmt.Errorf("%w failed to send request for location data", err)
	}

	timber.Debug("activity:", stravaActivity.Name)
	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no location results returned for %s", stravaActivity.Name)
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
		timber.Warning("unable to create location for", stravaActivity.Name, fmt.Sprintf("(%f, %f)", latitude, longitude))
		return nil, nil
	}

	location = strings.TrimSpace(location)

	return &location, nil
}

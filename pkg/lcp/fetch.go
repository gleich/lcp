package lcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Cache string

const (
	AppleMusic Cache = "applemusic"
	GitHub     Cache = "github"
	Steam      Cache = "steam"
	Strava     Cache = "strava"
)

type Client struct {
	Token      string
	httpClient http.Client
}

type Response[T CacheData] struct {
	Data    T
	Updated time.Time
}

func FetchCache[T CacheData](client Client, cache Cache) (Response[T], error) {
	var zeroValue Response[T] // acts as "nil" value to be used when returning an error
	if client.Token == "" {
		return zeroValue, errors.New("no token provided in client")
	}

	url, err := url.JoinPath("https://lcp.dev.mattglei.ch", string(cache))
	if err != nil {
		return zeroValue, fmt.Errorf("%v failed to join path for URL", err)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return zeroValue, fmt.Errorf("%v failed to create request", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.Token))

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return zeroValue, fmt.Errorf("%v failed to execute request", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zeroValue, fmt.Errorf("%v reading request body failed", err)
	}

	var response Response[T]
	err = json.Unmarshal(body, &response)
	if err != nil {
		return zeroValue, fmt.Errorf("%v failed to parse json", err)
	}
	return response, nil
}

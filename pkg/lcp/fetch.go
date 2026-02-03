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

type Client struct {
	Token      string
	httpClient http.Client
}

type CacheResponse[T any] struct {
	Data    T         `json:"data"`
	Updated time.Time `json:"updated"`
}

func fetch[T any](client *Client, path string) (T, error) {
	var zero T // acts as "nil" value to be used when returning an error

	if client.Token == "" {
		return zero, errors.New("no token provided in client")
	}

	url, err := url.JoinPath("https://lcp.mattglei.ch", path)
	if err != nil {
		return zero, fmt.Errorf("joining url: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return zero, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.Token))

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("making requet: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return zero, fmt.Errorf("reading request body: %w", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return zero, fmt.Errorf("closing request body: %w", err)
	}

	var response T
	err = json.Unmarshal(body, &response)
	if err != nil {
		return zero, fmt.Errorf("parsing json: %w", err)
	}

	return response, nil
}

func FetchCache[T CacheResponseData](client *Client) (CacheResponse[T], error) {
	var zero CacheResponse[T] // acts as "nil" value to be used when returning an error

	var cacheName string
	switch any(zero.Data).(type) {
	case AppleMusicCache:
		cacheName = "applemusic"
	case []GitHubRepository:
		cacheName = "github"
	case []SteamGame:
		cacheName = "steam"
	case []Workout:
		cacheName = "workouts"
	}

	resp, err := fetch[CacheResponse[T]](client, cacheName)
	if err != nil {
		return zero, fmt.Errorf("%w failed to fetch data", err)
	}
	return resp, nil
}

func FetchAppleMusicSyncedPlaylists(client *Client) ([]AppleMusicSyncedPlaylist, error) {
	resp, err := fetch[[]AppleMusicSyncedPlaylist](client, "applemusic/playlists")
	if err != nil {
		return []AppleMusicSyncedPlaylist{}, fmt.Errorf("%w failed to fetch data", err)
	}
	return resp, nil
}

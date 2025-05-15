package steam

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"go.mattglei.ch/lcp/internal/apis"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/pkg/lcp"
)

type recentlyPlayedResponse struct {
	Response struct {
		Games []struct {
			Name            string `json:"name"`
			AppID           int32  `json:"appid"`
			ImgIconURL      string `json:"img_icon_url"`
			PlaytimeForever int32  `json:"playtime_forever"`
		} `json:"games"`
	} `json:"response"`
}

type lastPlayedTimesResponse struct {
	Response struct {
		Games []struct {
			AppID        int32 `json:"appid"`
			LastPlaytime int64 `json:"last_playtime"`
		}
	}
}

func fetchRecentlyPlayedGames(client *http.Client) ([]lcp.SteamGame, error) {
	params := url.Values{
		"key":     {secrets.ENV.SteamKey},
		"steamid": {secrets.ENV.SteamID},
		"format":  {"json"},
	}
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.steampowered.com/IPlayerService/GetRecentlyPlayedGames/v0001/?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%w failed to create request for steam API owned games", err)
	}
	recentlyPlayedGames, err := apis.RequestJSON[recentlyPlayedResponse](
		cacheInstance.LogPrefix(),
		client,
		req,
	)
	if err != nil {
		return nil, fmt.Errorf("%w sending request for owned games failed", err)
	}

	// undocumented API access being used here. could potentially cause problems due to pagination
	// especially when I own over 100 games
	params = url.Values{
		"key": {secrets.ENV.SteamKey},
	}
	req, err = http.NewRequest(
		http.MethodGet,
		"https://api.steampowered.com/IPlayerService/ClientGetLastPlayedTimes/v1/?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%w failed to create request for steam API last played games", err)
	}
	lastPlayedTimes, err := apis.RequestJSON[lastPlayedTimesResponse](
		cacheInstance.LogPrefix(),
		client,
		req,
	)
	if err != nil {
		return nil, fmt.Errorf("%w sending request for steam API last played games failed", err)
	}

	var games []lcp.SteamGame
	for _, g := range recentlyPlayedGames.Response.Games {
		achievementPercentage, achievements, err := fetchGameAchievements(client, g.AppID)
		if err != nil {
			return nil, err
		}

		lastPlaytime := int64(0)
		for _, game := range lastPlayedTimes.Response.Games {
			if g.AppID == game.AppID {
				lastPlaytime = game.LastPlaytime
			}
		}

		games = append(games, lcp.SteamGame{
			Name:  g.Name,
			AppID: g.AppID,
			IconURL: fmt.Sprintf(
				"https://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg",
				g.AppID,
				g.ImgIconURL,
			),
			RTimeLastPlayed: time.Unix(lastPlaytime, 0),
			PlaytimeForever: g.PlaytimeForever,
			URL:             fmt.Sprintf("https://store.steampowered.com/app/%d/", g.AppID),
			HeaderURL: fmt.Sprintf(
				"https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/%d/header.jpg",
				g.AppID,
			),
			LibraryHeroURL: fmt.Sprintf(
				"https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/%d/library_hero.jpg",
				g.AppID,
			),
			LibraryHeroLogoURL: fmt.Sprintf(
				"https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/%d/logo.png",
				g.AppID,
			),
			AchievementProgress: achievementPercentage,
			Achievements:        achievements,
		})

	}

	sort.Slice(games, func(i, j int) bool {
		return games[j].RTimeLastPlayed.Before(games[i].RTimeLastPlayed)
	})

	return games, nil
}

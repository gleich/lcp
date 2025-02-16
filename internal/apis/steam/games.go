package steam

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"go.mattglei.ch/lcp-2/internal/apis"
	"go.mattglei.ch/lcp-2/internal/secrets"
	"go.mattglei.ch/lcp-2/pkg/lcp"
)

type ownedGamesResponse struct {
	Response struct {
		Games []struct {
			Name            string `json:"name"`
			AppID           int32  `json:"appid"`
			ImgIconURL      string `json:"img_icon_url"`
			RTimeLastPlayed int64  `json:"rtime_last_played"`
			PlaytimeForever int32  `json:"playtime_forever"`
		} `json:"games"`
	} `json:"response"`
}

func fetchRecentlyPlayedGames(client *http.Client) ([]lcp.SteamGame, error) {
	params := url.Values{
		"key":             {secrets.ENV.SteamKey},
		"steamid":         {secrets.ENV.SteamID},
		"include_appinfo": {"true"},
		"format":          {"json"},
	}
	req, err := http.NewRequest(http.MethodGet,
		"https://api.steampowered.com/IPlayerService/GetOwnedGames/v1?"+params.Encode(), nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%w failed to create request for steam API owned games", err)
	}
	ownedGames, err := apis.SendRequest[ownedGamesResponse](client, req)
	if err != nil {
		if !errors.Is(err, apis.IgnoreError) {
			return nil, fmt.Errorf("%w sending request for owned games failed", err)
		}
		return nil, err
	}

	sort.Slice(ownedGames.Response.Games, func(i, j int) bool {
		return ownedGames.Response.Games[i].RTimeLastPlayed > ownedGames.Response.Games[j].RTimeLastPlayed
	})

	var games []lcp.SteamGame
	i := 0
	for len(games) < 10 {
		if i > len(games) {
			break
		}
		g := ownedGames.Response.Games[i]
		i++
		libraryURL := fmt.Sprintf(
			"https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/%d/library_600x900.jpg",
			g.AppID,
		)
		libraryImageResponse, err := http.Get(libraryURL)
		if err != nil {
			return nil, fmt.Errorf("%w getting library image for %s failed", err, g.Name)
		}
		defer libraryImageResponse.Body.Close()

		var libraryURLPtr *string
		if libraryImageResponse.StatusCode == http.StatusOK {
			libraryURLPtr = &libraryURL
		}

		achievementPercentage, achievements, err := fetchGameAchievements(client, g.AppID)
		if err != nil {
			return nil, err
		}

		games = append(games, lcp.SteamGame{
			Name:  g.Name,
			AppID: g.AppID,
			IconURL: fmt.Sprintf(
				"https://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg",
				g.AppID,
				g.ImgIconURL,
			),
			RTimeLastPlayed: time.Unix(g.RTimeLastPlayed, 0),
			PlaytimeForever: g.PlaytimeForever,
			URL:             fmt.Sprintf("https://store.steampowered.com/app/%d/", g.AppID),
			HeaderURL: fmt.Sprintf(
				"https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/%d/header.jpg",
				g.AppID,
			),
			LibraryURL: libraryURLPtr,
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

	return games, nil
}

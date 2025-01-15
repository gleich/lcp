package steam

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"pkg.mattglei.ch/lcp-2/internal/apis"
	"pkg.mattglei.ch/lcp-2/internal/secrets"
	"pkg.mattglei.ch/timber"
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

type game struct {
	Name                string         `json:"name"`
	AppID               int32          `json:"app_id"`
	IconURL             string         `json:"icon_url"`
	RTimeLastPlayed     time.Time      `json:"rtime_last_played"`
	PlaytimeForever     int32          `json:"playtime_forever"`
	URL                 string         `json:"url"`
	HeaderURL           string         `json:"header_url"`
	LibraryURL          *string        `json:"library_url"`
	LibraryHeroURL      string         `json:"library_hero_url"`
	LibraryHeroLogoURL  string         `json:"library_hero_logo_url"`
	AchievementProgress *float32       `json:"achievement_progress"`
	Achievements        *[]achievement `json:"achievements"`
}

func fetchRecentlyPlayedGames() ([]game, error) {
	params := url.Values{
		"key":             {secrets.SECRETS.SteamKey},
		"steamid":         {secrets.SECRETS.SteamID},
		"include_appinfo": {"true"},
		"format":          {"json"},
	}
	req, err := http.NewRequest(http.MethodGet,
		"https://api.steampowered.com/IPlayerService/GetOwnedGames/v1?"+params.Encode(), nil,
	)
	if err != nil {
		timber.Error(err, "failed to create request for steam API owned games")
		return nil, err
	}
	ownedGames, err := apis.SendRequest[ownedGamesResponse](req)
	if err != nil {
		if !errors.Is(err, apis.WarningError) {
			timber.Error(err, "sending request for owned games failed")
		}
		return nil, err
	}

	sort.Slice(ownedGames.Response.Games, func(i, j int) bool {
		return ownedGames.Response.Games[i].RTimeLastPlayed > ownedGames.Response.Games[j].RTimeLastPlayed
	})

	var games []game
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
			timber.Error(err, "getting library image for", g.Name, "failed")
			return nil, err
		}
		defer libraryImageResponse.Body.Close()

		var libraryURLPtr *string
		if libraryImageResponse.StatusCode == http.StatusOK {
			libraryURLPtr = &libraryURL
		}

		achievementPercentage, achievements, err := fetchGameAchievements(g.AppID)
		if err != nil {
			return nil, err
		}

		games = append(games, game{
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

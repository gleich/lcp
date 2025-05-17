package steam

import (
	"fmt"
	"image/jpeg"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/apis"
	"go.mattglei.ch/lcp/internal/images"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/pkg/lcp"
)

type ownedGamesResponse struct {
	Response struct {
		Games []struct {
			Name            string `json:"name"`
			AppID           int32  `json:"appid"`
			LastPlayed      int64  `json:"rtime_last_played"`
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

func fetchRecentlyPlayedGames(client *http.Client, rdb *redis.Client) ([]lcp.SteamGame, error) {
	params := url.Values{
		"key":             {secrets.ENV.SteamKey},
		"steamid":         {secrets.ENV.SteamID},
		"include_appinfo": {"true"},
	}
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.steampowered.com/IPlayerService/GetOwnedGames/v1/?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%w failed to create request for steam API owned games", err)
	}
	ownedGames, err := apis.RequestJSON[ownedGamesResponse](
		cacheInstance.LogPrefix(),
		client,
		req,
	)
	if err != nil {
		return nil, fmt.Errorf("%w sending request for owned games failed", err)
	}

	sort.Slice(ownedGames.Response.Games, func(i, j int) bool {
		return ownedGames.Response.Games[j].LastPlayed < ownedGames.Response.Games[i].LastPlayed
	})

	var games []lcp.SteamGame
	for _, g := range ownedGames.Response.Games[:10] {
		achievementPercentage, achievements, err := fetchGameAchievements(client, g.AppID)
		if err != nil {
			return nil, err
		}

		headerURL := fmt.Sprintf(
			"https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/%d/header.jpg",
			g.AppID,
		)
		headerBlurHash, err := images.BlurHash(client, rdb, headerURL, jpeg.Decode)
		if err != nil {
			return nil, fmt.Errorf("%w failed to load blurhash image data for library hero", err)
		}

		games = append(games, lcp.SteamGame{
			Name:  g.Name,
			AppID: g.AppID,
			IconURL: fmt.Sprintf(
				"https://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg",
				g.AppID,
				g.ImgIconURL,
			),
			RTimeLastPlayed: time.Unix(g.LastPlayed, 0),
			PlaytimeForever: g.PlaytimeForever,
			URL:             fmt.Sprintf("https://store.steampowered.com/app/%d/", g.AppID),
			HeaderURL: fmt.Sprintf(
				"https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/%d/header.jpg",
				g.AppID,
			),
			HeaderBlurHash: headerBlurHash,
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

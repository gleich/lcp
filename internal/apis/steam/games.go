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
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/images"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/pkg/lcp"
)

type ownedGamesResponse struct {
	Response struct {
		Apps []struct {
			AppID        int    `json:"appid"`
			Name         string `json:"name"`
			RtLastPlayed int    `json:"rt_last_played"`
			RtPlaytime   int    `json:"rt_playtime"`
		} `json:"apps"`
	} `json:"response"`
}

func fetchRecentlyPlayedGames(client *http.Client, rdb *redis.Client) ([]lcp.SteamGame, error) {
	params := url.Values{
		"access_token":      {secrets.ENV.SteamWebAjaxToken},
		"family_groupid":    {"0"},
		"include_own":       {"true"},
		"include_excluded":  {"true"},
		"include_free":      {"true"},
		"include_non_games": {"false"},
	}
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.steampowered.com/IFamilyGroupsService/GetSharedLibraryApps/v1/?"+params.Encode(),
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

	sort.Slice(ownedGames.Response.Apps, func(i, j int) bool {
		return ownedGames.Response.Apps[j].RtLastPlayed < ownedGames.Response.Apps[i].RtLastPlayed
	})

	if len(ownedGames.Response.Apps) < 10 {
		return nil, cache.ErrSteamOwnedGamesEmpty
	}

	var games []lcp.SteamGame
	for _, g := range ownedGames.Response.Apps[:10] {
		achievementPercentage, err := fetchAchievementsPercentage(client, g.AppID)
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
			Name:            g.Name,
			AppID:           g.AppID,
			RTimeLastPlayed: time.Unix(int64(g.RtLastPlayed), 0),
			PlaytimeForever: g.RtPlaytime,
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
		})

	}

	return games, nil
}

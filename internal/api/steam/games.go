package steam

import (
	"cmp"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/images"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/pkg/lcp"
)

const storeItemAssetsBaseURL = "https://shared.akamai.steamstatic.com/store_item_assets/"

type ownedGamesResponse struct {
	Response struct {
		Games []struct {
			Name                     string `json:"name"`
			AppID                    int    `json:"appid"`
			LastPlayed               int64  `json:"rtime_last_played"`
			ImgIconURL               string `json:"img_icon_url"`
			PlaytimeForever          int    `json:"playtime_forever"`
			HasCommunityVisibleStats bool   `json:"has_community_visible_stats"`
		} `json:"games"`
	} `json:"response"`
}

type storeItemsResponse struct {
	Response struct {
		StoreItems []struct {
			AppID  int             `json:"appid"`
			Assets storeItemAssets `json:"assets"`
		} `json:"store_items"`
	} `json:"response"`
}

// storeItemAssets holds the versioned asset paths Steam returns for a store item. Newer apps
// have a content hash in the filename (e.g. "803663.../header.jpg") while older apps return a
// bare filename, so the unhashed convenience path can 404. AssetURLFormat is a template like
// "steam/apps/4121170/${FILENAME}?t=1781141362" that we fill in per asset.
type storeItemAssets struct {
	AssetURLFormat string `json:"asset_url_format"`
	Header         string `json:"header"`
	LibraryHero    string `json:"library_hero"`
}

func (a storeItemAssets) assetURL(appID int, filename string) string {
	format := cmp.Or(a.AssetURLFormat, fmt.Sprintf("steam/apps/%d/${FILENAME}", appID))
	return storeItemAssetsBaseURL + strings.Replace(format, "${FILENAME}", filename, 1)
}

func fetchStoreItemAssets(client *http.Client, appIDs []int) (map[int]storeItemAssets, error) {
	ids := make([]map[string]int, len(appIDs))
	for i, id := range appIDs {
		ids[i] = map[string]int{"appid": id}
	}
	inputJSON, err := json.Marshal(map[string]any{
		"ids": ids,
		"context": map[string]string{
			"language":     "english",
			"country_code": "US",
		},
		"data_request": map[string]bool{
			"include_assets": true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("encoding store items input json: %w", err)
	}

	params := url.Values{"input_json": {string(inputJSON)}}
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.steampowered.com/IStoreBrowseService/GetItems/v1/?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request for steam store items: %w", err)
	}
	storeItems, err := api.RequestJSON[storeItemsResponse](client, req, logger())
	if err != nil {
		return nil, fmt.Errorf("sending request for store items: %w", err)
	}

	assets := make(map[int]storeItemAssets, len(storeItems.Response.StoreItems))
	for _, item := range storeItems.Response.StoreItems {
		assets[item.AppID] = item.Assets
	}
	return assets, nil
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
		return nil, fmt.Errorf("creating request for steam API owned games: %w", err)
	}
	ownedGames, err := api.RequestJSON[ownedGamesResponse](client, req, logger())
	if err != nil {
		return nil, fmt.Errorf("sending request for owned games: %w", err)
	}

	sort.Slice(ownedGames.Response.Games, func(i, j int) bool {
		return ownedGames.Response.Games[j].LastPlayed < ownedGames.Response.Games[i].LastPlayed
	})

	if len(ownedGames.Response.Games) < 6 {
		return nil, cache.ErrSteamOwnedGamesEmpty
	}

	topGames := ownedGames.Response.Games[:10]
	appIDs := make([]int, len(topGames))
	for i, g := range topGames {
		appIDs[i] = g.AppID
	}
	assets, err := fetchStoreItemAssets(client, appIDs)
	if err != nil {
		return nil, err
	}

	var games []lcp.SteamGame
	for _, g := range topGames {
		gameAssets := assets[g.AppID]
		game := lcp.SteamGame{
			Name:  g.Name,
			AppID: g.AppID,
			IconURL: fmt.Sprintf(
				"https://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg",
				g.AppID,
				g.ImgIconURL,
			),
			RTimeLastPlayed:    time.Unix(g.LastPlayed, 0),
			PlaytimeForever:    g.PlaytimeForever,
			URL:                fmt.Sprintf("https://store.steampowered.com/app/%d/", g.AppID),
			HeaderURL:          gameAssets.assetURL(g.AppID, cmp.Or(gameAssets.Header, "header.jpg")),
			LibraryHeroURL:     gameAssets.assetURL(g.AppID, cmp.Or(gameAssets.LibraryHero, "library_hero.jpg")),
			LibraryHeroLogoURL: gameAssets.assetURL(g.AppID, "logo.png"),
		}

		if g.HasCommunityVisibleStats {
			achievementPercentage, err := fetchAchievementsPercentage(client, g.AppID)
			if err != nil {
				return nil, err
			}
			game.AchievementProgress = achievementPercentage
		}

		headerBlurHash, err := images.BlurHash(client, rdb, game.HeaderURL, jpeg.Decode, logger())
		if err != nil {
			logger().Warn().Err(err).Int("app_id", g.AppID).Msg("failed to generate header blurhash")
		} else {
			game.HeaderBlurHash = headerBlurHash
		}

		games = append(games, game)
	}

	return games, nil
}

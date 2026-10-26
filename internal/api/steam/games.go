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

type lastPlayedTimesResponse struct {
	Response struct {
		Games []struct {
			AppID           int   `json:"appid"`
			LastPlaytime    int64 `json:"last_playtime"`
			PlaytimeForever int   `json:"playtime_forever"`
		} `json:"games"`
	} `json:"response"`
}

type storeItem struct {
	AppID  int             `json:"appid"`
	Name   string          `json:"name"`
	Assets storeItemAssets `json:"assets"`
}

type storeItemsResponse struct {
	Response struct {
		StoreItems []storeItem `json:"store_items"`
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
	CommunityIcon  string `json:"community_icon"`
}

func (a storeItemAssets) assetURL(appID int, filename string) string {
	format := cmp.Or(a.AssetURLFormat, fmt.Sprintf("steam/apps/%d/${FILENAME}", appID))
	return storeItemAssetsBaseURL + strings.Replace(format, "${FILENAME}", filename, 1)
}

func fetchStoreItems(client *http.Client, appIDs []int) (map[int]storeItem, error) {
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

	items := make(map[int]storeItem, len(storeItems.Response.StoreItems))
	for _, item := range storeItems.Response.StoreItems {
		items[item.AppID] = item
	}
	return items, nil
}

// fetchRecentlyPlayedGames uses ClientGetLastPlayedTimes instead of GetOwnedGames because it
// also includes games borrowed through Steam Family Sharing. It has no steamid parameter and
// always returns the games for the owner of the API key.
func fetchRecentlyPlayedGames(client *http.Client, rdb *redis.Client) ([]lcp.SteamGame, error) {
	params := url.Values{"key": {secrets.ENV.SteamKey}}
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.steampowered.com/IPlayerService/ClientGetLastPlayedTimes/v1/?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request for steam API last played times: %w", err)
	}
	lastPlayedTimes, err := api.RequestJSON[lastPlayedTimesResponse](client, req, logger())
	if err != nil {
		return nil, fmt.Errorf("sending request for last played times: %w", err)
	}

	playedGames := lastPlayedTimes.Response.Games
	sort.Slice(playedGames, func(i, j int) bool {
		return playedGames[j].LastPlaytime < playedGames[i].LastPlaytime
	})

	if len(playedGames) < 6 {
		return nil, cache.ErrSteamOwnedGamesEmpty
	}

	topGames := playedGames[:10]
	appIDs := make([]int, len(topGames))
	for i, g := range topGames {
		appIDs[i] = g.AppID
	}
	storeItems, err := fetchStoreItems(client, appIDs)
	if err != nil {
		return nil, err
	}

	var games []lcp.SteamGame
	for _, g := range topGames {
		item := storeItems[g.AppID]
		if item.Name == "" {
			logger().Warn().Int("app_id", g.AppID).Msg("no store item found for game")
			continue
		}

		game := lcp.SteamGame{
			Name:  item.Name,
			AppID: g.AppID,
			IconURL: fmt.Sprintf(
				"https://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg",
				g.AppID,
				item.Assets.CommunityIcon,
			),
			RTimeLastPlayed:    time.Unix(g.LastPlaytime, 0),
			PlaytimeForever:    g.PlaytimeForever,
			URL:                fmt.Sprintf("https://store.steampowered.com/app/%d/", g.AppID),
			HeaderURL:          item.Assets.assetURL(g.AppID, cmp.Or(item.Assets.Header, "header.jpg")),
			LibraryHeroURL:     item.Assets.assetURL(g.AppID, cmp.Or(item.Assets.LibraryHero, "library_hero.jpg")),
			LibraryHeroLogoURL: item.Assets.assetURL(g.AppID, "logo.png"),
		}

		achievementPercentage, err := fetchAchievementsPercentage(client, g.AppID)
		if err != nil {
			return nil, err
		}
		game.AchievementProgress = achievementPercentage

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

package steam

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"go.mattglei.ch/lcp/internal/apis"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

type playerAchievementsResponse struct {
	PlayerStats struct {
		Achievements *[]struct {
			ApiName    string `json:"apiname"`
			Achieved   int    `json:"achieved"`
			UnlockTime *int64 `json:"unlocktime"`
		}
	} `json:"playerStats"`
}

type schemaGameResponse struct {
	Game struct {
		GameStats struct {
			Achievements []struct {
				DisplayName string  `json:"displayName"`
				Icon        string  `json:"icon"`
				Description *string `json:"description"`
				Name        string  `json:"name"`
			} `json:"achievements"`
		} `json:"availableGameStats"`
	} `json:"game"`
}

func fetchGameAchievements(
	client *http.Client,
	appID int32,
) (*float32, *[]lcp.SteamAchievement, error) {
	params := url.Values{
		"key":     {secrets.ENV.SteamKey},
		"steamid": {secrets.ENV.SteamID},
		"appid":   {fmt.Sprint(appID)},
		"format":  {"json"},
	}
	resp, err := client.Get(
		"https://api.steampowered.com/ISteamUserStats/GetPlayerAchievements/v0001?" + params.Encode(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"%v sending request for player achievements from %d failed",
			err,
			appID,
		)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"%v reading response body for player achievements from %d failed",
			err,
			appID,
		)
	}
	if string(body) == `{"playerstats":{"error":"Requested app has no stats","success":false}}` {
		return nil, nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		timber.Warning(
			"status code of",
			resp.StatusCode,
			"returned from API. Code of 200 expected from",
			resp.Request.URL.String(),
		)
		return nil, nil, apis.ErrWarning
	}

	var playerAchievements playerAchievementsResponse
	err = json.Unmarshal(body, &playerAchievements)
	if err != nil {
		err = fmt.Errorf("%w failed to parse json for player achievements for %d", err, appID)
		timber.Debug("body:", string(body))
		return nil, nil, err
	}

	if playerAchievements.PlayerStats.Achievements == nil {
		return nil, nil, nil
	}

	params = url.Values{
		"key":    {secrets.ENV.SteamKey},
		"appid":  {fmt.Sprint(appID)},
		"format": {"json"},
	}
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.steampowered.com/ISteamUserStats/GetSchemaForGame/v2?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"%v creating request for owned games failed for app id: %d",
			err,
			appID,
		)
	}
	gameSchema, err := apis.RequestJSON[schemaGameResponse](cacheInstance.LogPrefix(), client, req)
	if err != nil {
		return nil, nil, fmt.Errorf("%w failed to get game schema for app id: %d", err, appID)
	}

	var achievements []lcp.SteamAchievement
	for _, playerAchievement := range *playerAchievements.PlayerStats.Achievements {
		for _, schemaAchievement := range gameSchema.Game.GameStats.Achievements {
			if playerAchievement.ApiName == schemaAchievement.Name {
				var unlockTime time.Time
				if playerAchievement.UnlockTime != nil && *playerAchievement.UnlockTime != 0 {
					unlockTime = time.Unix(*playerAchievement.UnlockTime, 0)
				}
				achievements = append(achievements, lcp.SteamAchievement{
					ApiName:     playerAchievement.ApiName,
					Achieved:    playerAchievement.Achieved == 1,
					Icon:        schemaAchievement.Icon,
					DisplayName: schemaAchievement.DisplayName,
					Description: schemaAchievement.Description,
					UnlockTime:  &unlockTime,
				})
			}
		}
	}

	var totalAchieved int
	for _, achievement := range achievements {
		if achievement.Achieved {
			totalAchieved++
		}
	}
	achievementPercentage := (float32(totalAchieved) / float32(len(achievements))) * 100.0

	sort.Slice(achievements, func(i, j int) bool {
		if achievements[i].UnlockTime == nil && achievements[j].UnlockTime == nil {
			return false
		}
		if achievements[i].UnlockTime == nil {
			return false
		}
		if achievements[j].UnlockTime == nil {
			return true
		}
		return achievements[i].UnlockTime.After(*achievements[j].UnlockTime)
	})

	if len(achievements) > 5 {
		achievements = achievements[:5]
	}

	return &achievementPercentage, &achievements, nil
}

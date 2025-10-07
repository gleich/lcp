package steam

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.mattglei.ch/lcp/internal/apis"
	"go.mattglei.ch/lcp/internal/secrets"
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

type achievement struct {
	Achieved bool `json:"achieved"`
}

func fetchAchievementsPercentage(
	client *http.Client,
	appID int,
) (*float32, error) {
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
		return nil, fmt.Errorf(
			"%v sending request for player achievements from %d failed",
			err,
			appID,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"%v reading response body for player achievements from %d failed",
			err,
			appID,
		)
	}
	if string(body) == `{"playerstats":{"error":"Requested app has no stats","success":false}}` {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		timber.Warning(
			cacheInstance.LogPrefix(),
			"status code of",
			resp.StatusCode,
			"returned from API. Code of 200 expected from",
			resp.Request.URL.String(),
		)
		return nil, apis.ErrWarning
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("%w failed to close response body", err)
	}

	var playerAchievements playerAchievementsResponse
	err = json.Unmarshal(body, &playerAchievements)
	if err != nil {
		err = fmt.Errorf("%w failed to parse json for player achievements for %d", err, appID)
		timber.Debug("body:", string(body))
		return nil, err
	}

	if playerAchievements.PlayerStats.Achievements == nil {
		return nil, nil
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
		return nil, fmt.Errorf(
			"%v creating request for owned games failed for app id: %d",
			err,
			appID,
		)
	}
	gameSchema, err := apis.RequestJSON[schemaGameResponse](cacheInstance.LogPrefix(), client, req)
	if err != nil {
		return nil, fmt.Errorf("%w failed to get game schema for app id: %d", err, appID)
	}

	var achievements []achievement
	for _, playerAchievement := range *playerAchievements.PlayerStats.Achievements {
		for _, schemaAchievement := range gameSchema.Game.GameStats.Achievements {
			if playerAchievement.ApiName == schemaAchievement.Name {
				achievements = append(achievements, achievement{
					Achieved: playerAchievement.Achieved == 1,
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

	return &achievementPercentage, nil
}

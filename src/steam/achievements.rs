use std::env;

use anyhow::{Context, Result};
use reqwest::Client;
use serde::{Deserialize, Serialize};

use super::{STEAM_ID, STEAM_TOKEN};

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct SchemaForGame {
    game: Game,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct Game {
    #[serde(rename = "availableGameStats")]
    available_game_stats: GameStats,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct GameStats {
    achievements: Vec<SchemaAchievement>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct SchemaAchievement {
    #[serde(rename = "displayName")]
    display_name: String,
    icon: String,
    description: String,
    name: String,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct PlayerAchievements {
    #[serde(rename = "playerstats")]
    player_stats: PlayerStats,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct PlayerStats {
    achievements: Option<Vec<PlayerAchievement>>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct PlayerAchievement {
    #[serde(rename = "apiname")]
    pub api_name: String,
    pub achieved: u32,
}

#[derive(Debug, PartialEq, Clone, Deserialize, Serialize)]
pub struct Achievement {
    pub api_name: String,
    pub achieved: bool,
    pub icon: String,
    pub display_name: String,
    pub description: String,
}

pub async fn fetch_game_achievements(
    app_id: u32,
    client: &Client,
) -> Result<Option<Vec<Achievement>>> {
    let steam_token = env::var(STEAM_TOKEN).context("fetching steam token env var failed")?;
    let steam_id = env::var(STEAM_ID).context("fetching steam id env var failed")?;

    let player_resp: reqwest::Response = client
        .get("https://api.steampowered.com/ISteamUserStats/GetPlayerAchievements/v0001/")
        .query(&[
            ("key", &steam_token),
            ("steamid", &steam_id),
            ("appid", &app_id.to_string()),
            ("format", &String::from("json")),
        ])
        .send()
        .await
        .context("getting response for player achievements failed")?;
    let player_resp_text = player_resp
        .text()
        .await
        .context("getting raw response text failed")?;
    if player_resp_text
        == r#"{"playerstats":{"error":"Requested app has no stats","success":false}}"#
    {
        return Ok(None);
    }
    let player_data: PlayerAchievements =
        serde_json::from_str(&player_resp_text).context(format!(
            "reading json failed from request to get achievements for {}: response: {}",
            app_id, player_resp_text
        ))?;
    let player_achievements = player_data.player_stats.achievements;
    if player_achievements.is_none() {
        return Ok(None);
    }

    let schema_resp: reqwest::Response = client
        .get("https://api.steampowered.com/ISteamUserStats/GetSchemaForGame/v2/")
        .query(&[
            ("key", &steam_token),
            ("appid", &app_id.to_string()),
            ("format", &String::from("json")),
        ])
        .send()
        .await
        .context("getting response for player achievements failed")?;
    let schema_resp_text = schema_resp
        .text()
        .await
        .context("getting raw response text failed")?;
    let schema_data: SchemaForGame = serde_json::from_str(&schema_resp_text).context(format!(
        "reading json failed from request to get achievements for {}: response: {}",
        app_id, schema_resp_text
    ))?;

    let mut achievements = vec![];
    for player_achievement in player_achievements.unwrap() {
        for schema_achievement in &schema_data.game.available_game_stats.achievements {
            if player_achievement.api_name == schema_achievement.name {
                achievements.push(Achievement {
                    api_name: player_achievement.api_name.to_owned(),
                    achieved: player_achievement.achieved == 1,
                    icon: schema_achievement.icon.to_owned(),
                    display_name: schema_achievement.display_name.to_owned(),
                    description: schema_achievement.description.to_owned(),
                })
            }
        }
    }

    Ok(Some(achievements))
}

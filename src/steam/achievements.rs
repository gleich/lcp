use std::env;

use anyhow::{Context, Result};
use reqwest::Client;
use serde::{Deserialize, Serialize};

use super::{STEAM_ID, STEAM_TOKEN};

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct MainResponse {
    #[serde(rename = "playerstats")]
    player_stats: PlayerStats,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct PlayerStats {
    achievements: Vec<Achievement>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct Achievement {
    #[serde(rename = "apiname")]
    pub api_name: String,
    pub achieved: u32,
}

pub async fn fetch_game_achievements(
    app_id: u32,
    client: &Client,
) -> Result<Option<Vec<Achievement>>> {
    let resp: reqwest::Response = client
        .get("https://api.steampowered.com/ISteamUserStats/GetPlayerAchievements/v0001/")
        .query(&[
            (
                "key",
                env::var(STEAM_TOKEN).context("fetching steam token env var failed")?,
            ),
            (
                "steamid",
                env::var(STEAM_ID).context("fetching steam id env var failed")?,
            ),
            ("appid", app_id.to_string()),
            ("format", String::from("json")),
        ])
        .send()
        .await
        .context("getting response for activity failed")?;
    let resp_text = resp
        .text()
        .await
        .context("getting raw response text failed")?;
    if resp_text == r#"{"playerstats":{"error":"Requested app has no stats","success":false}}"# {
        return Ok(None);
    }
    let data: MainResponse = serde_json::from_str(&resp_text).context(format!(
        "reading json failed from request to get achievements for {}: response: {}",
        app_id, resp_text
    ))?;
    Ok(Some(data.player_stats.achievements))
}

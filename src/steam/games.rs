use std::env;

use anyhow::{Context, Result};
use chrono::{DateTime, TimeZone, Utc};
use reqwest::Client;
use serde::{Deserialize, Serialize};

const STEAM_TOKEN: &str = "STEAM_ACCESS_TOKEN";
const STEAM_ID: &str = "STEAM_ID";

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct MainSteamResponse {
    pub response: SteamResponse,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct SteamResponse {
    pub games: Vec<RawGame>,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct RawGame {
    pub name: String,
    pub appid: u32,
    pub img_icon_url: String,
    pub rtime_last_played: i64,
    pub playtime_forever: u32,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Game {
    pub name: String,
    pub app_id: u32,
    pub url: String,
    pub icon_url: String,
    pub header_url: String,
    pub library_url: String,
    pub playtime_forever: u32,
    pub rtime_last_played: DateTime<Utc>,
}

pub async fn fetch_recently_played(client: &Client) -> Result<Vec<Game>> {
    let mut resp: MainSteamResponse = client
        .get("https://api.steampowered.com/IPlayerService/GetOwnedGames/v1/")
        .query(&[
            (
                "key",
                env::var(STEAM_TOKEN).context("fetching steam token env var failed")?,
            ),
            (
                "steamid",
                env::var(STEAM_ID).context("fetching steam id env var failed")?,
            ),
            ("include_appinfo", String::from("true")),
            ("format", String::from("json")),
        ])
        .send()
        .await
        .context("sending request for recent games failed")?
        .json()
        .await
        .context("reading json failed from request to get recent games")?;
    let mut games: Vec<Game> = resp
        .response
        .games
        .iter_mut()
        .map(|g| Game {
            name: g.name.to_owned(),
            url: format!("https://store.steampowered.com/app/{}/", &g.appid),
            icon_url: format!(
                "http://media.steampowered.com/steamcommunity/public/images/apps/{}/{}.jpg",
                &g.appid, g.img_icon_url
            ),
            library_url: format!(
                "https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/{}/library_600x900.jpg",
                &g.appid,
            ),
            header_url: format!(
                "https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/{}/header.jpg",
                &g.appid,
            ),
            app_id: g.appid,
            playtime_forever: g.playtime_forever,
            rtime_last_played: Utc.timestamp_opt(g.rtime_last_played, 0).unwrap()
        })
        .collect();
    games.sort_by(|a, b| b.rtime_last_played.cmp(&a.rtime_last_played));
    Ok(games)
}

use std::env;

use anyhow::{Context, Result};
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
    pub playtime_2weeks: u32,
    pub playtime_forever: u32,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Game {
    pub name: String,
    pub app_id: u32,
    pub icon_url: String,
    pub playtime: Playtime,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Playtime {
    pub minutes_last_2weeks: u32,
    pub minutes_forever: u32,
}

pub async fn fetch_recent(client: &Client) -> Result<Vec<Game>> {
    let mut resp: MainSteamResponse = client
        .get("http://api.steampowered.com/IPlayerService/GetRecentlyPlayedGames/v0001/")
        .query(&[
            (
                "key",
                env::var(STEAM_TOKEN).context("fetching steam token env var failed")?,
            ),
            (
                "steamid",
                env::var(STEAM_ID).context("fetching steam id env var failed")?,
            ),
            ("format", String::from("json")),
        ])
        .send()
        .await
        .context("sending request for recent games failed")?
        .json()
        .await
        .context("reading json failed from request to get recent games")?;
    Ok(resp
        .response
        .games
        .iter_mut()
        .map(|g| Game {
            name: g.name.to_owned(),
            icon_url: format!(
                "http://media.steampowered.com/steamcommunity/public/images/apps/{}/{}.jpg",
                &g.appid, g.img_icon_url
            ),
            app_id: g.appid,
            playtime: Playtime {
                minutes_last_2weeks: g.playtime_2weeks,
                minutes_forever: g.playtime_forever,
            },
        })
        .collect())
}

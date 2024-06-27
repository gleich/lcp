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
    pub library_url: Option<String>,
    // pub playtime_forever: u32,
    // pub rtime_last_played: DateTime<Utc>,
}

pub async fn fetch_recently_played(client: &Client) -> Result<Vec<Game>> {
    let resp: reqwest::Response = client
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
        .context("sending request for recent games failed")?;
    let resp_text = resp
        .text()
        .await
        .context("getting raw response text failed")?;
    let mut data: MainSteamResponse = serde_json::from_str(&resp_text).context(format!(
        "reading json failed from request to get recent games: response: {}",
        resp_text
    ))?;
    data.response
        .games
        .sort_by(|a, b| b.rtime_last_played.cmp(&a.rtime_last_played));
    let mut games: Vec<Game> = vec![];
    for game in data.response.games.iter_mut() {
        let library_url = format!(
                "https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/{}/library_600x900.jpg",
                &game.appid,
            );
        let library_url_exists = client
            .get(&library_url)
            .send()
            .await
            .context(format!("checking library url for {} failed", &game.name))?
            .status()
            == 200;
        games.push(Game {
            name: game.name.to_owned(),
            url: format!("https://store.steampowered.com/app/{}/", &game.appid),
            icon_url: format!(
                "http://media.steampowered.com/steamcommunity/public/images/apps/{}/{}.jpg",
                &game.appid, game.img_icon_url
            ),
            library_url: if library_url_exists {
                Some(library_url)
            } else {
                None
            },
            header_url: format!(
                "https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/{}/header.jpg",
                &game.appid,
            ),
            app_id: game.appid,
            // playtime_forever: game.playtime_forever,
            // rtime_last_played: Utc.timestamp_opt(game.rtime_last_played, 0).unwrap(),
        })
    }
    Ok(games)
}

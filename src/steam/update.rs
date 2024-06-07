use std::{thread, time::Duration};

use anyhow::{Context, Result};
use reqwest::Client;
use tracing::info;

use super::{cache, games::fetch_recently_played};

pub async fn cache(client: &Client) -> Result<()> {
    let recently_played_games = fetch_recently_played(client)
        .await
        .context("fetching recently played games failed")?;
    cache::update(recently_played_games).expect("updating steam cache failed");
    Ok(())
}

pub async fn periodic_update() -> Result<()> {
    let client = Client::new();
    loop {
        cache(&client)
            .await
            .context("requesting recent games or updating cache failed")?;
        info!("polled steam API");
        thread::sleep(Duration::from_secs(60));
    }
}

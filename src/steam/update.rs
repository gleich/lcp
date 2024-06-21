use std::time::Duration;

use anyhow::{Context, Result};
use reqwest::Client;
use rocket::tokio;

use crate::reval::{self, Service};

use super::{cache, games::fetch_recently_played};

pub async fn cache(client: &Client) -> Result<()> {
    let recently_played_games = fetch_recently_played(client)
        .await
        .context("fetching recently played games failed")?;
    let updated_cache = cache::update(recently_played_games).expect("updating steam cache failed");
    if updated_cache {
        reval::call_for_revalidate(client, Service::Steam)
            .await
            .context("calling for website revalidation failed")?;
    }
    Ok(())
}

pub async fn periodic_update() -> Result<()> {
    let client = Client::new();
    loop {
        cache(&client)
            .await
            .context("requesting recent games or updating cache failed")?;
        tokio::time::sleep(Duration::from_secs(300)).await; // reload every 5 minutes
    }
}

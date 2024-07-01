use std::time::Duration;

use anyhow::{Context, Result};
use reqwest::Client;
use rocket::tokio;

use super::{cache, repos::fetch_pinned_repos};

pub async fn cache(client: &Client) -> Result<()> {
    let pinned_repos = fetch_pinned_repos(client)
        .await
        .context("fetching pinned repos failed")?;
    let updated_cache = cache::update(pinned_repos).expect("updating github cache failed");
    if updated_cache {
        // revalidation logic will go here in future
    }
    Ok(())
}

pub async fn periodic_update() -> Result<()> {
    let client = Client::new();
    loop {
        cache(&client)
            .await
            .context("requesting pinned repos or updating cache failed")?;
        tokio::time::sleep(Duration::from_secs(300)).await; // reload every 5 minutes
    }
}

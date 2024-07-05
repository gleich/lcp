use std::time::Duration;

use anyhow::{Context, Result};
use reqwest::Client;
use rocket::tokio;
use tracing::warn;
use tracing_error::SpanTrace;

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
        match cache(&client)
            .await
            .context("requesting recent games or updating cache failed")
        {
            Ok(()) => {}
            Err(err) => {
                let span_trace = SpanTrace::capture();
                warn!(
                    "encountered error trying to update cache: {}",
                    err.context(span_trace)
                );
            }
        }
        tokio::time::sleep(Duration::from_secs(300)).await; // reload every 5 minutes
    }
}

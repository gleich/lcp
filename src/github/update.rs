use std::time::Duration;

use anyhow::{Context, Result};
use reqwest::Client;
use rocket::tokio;
use tracing::warn;
use tracing_error::SpanTrace;

use super::{cache, repos::fetch_pinned_repos};

pub async fn cache(client: &Client) -> Result<()> {
    let pinned_repos = fetch_pinned_repos(client)
        .await
        .context("fetching pinned repos failed")?;
    cache::update(pinned_repos).expect("updating github cache failed");
    Ok(())
}

pub async fn periodic_update() -> Result<()> {
    let client = Client::new();
    loop {
        match cache(&client)
            .await
            .context("requesting pinned repos or updating cache failed")
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

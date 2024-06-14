use std::env;

use anyhow::{Context, Result};
use reqwest::Client;
use tracing::info;

const REVALIDATE_TOKEN: &str = "REVALIDATE_ACCESS_TOKEN";

pub async fn call_for_revalidate(client: &Client) -> Result<()> {
    client
        .post("https://beta.mattglei.ch/revalidate")
        .bearer_auth(env::var(REVALIDATE_TOKEN).context("getting revalidate token env var failed")?)
        .send()
        .await
        .context("sending request to revalidate website cache failed")?;
    info!("made call to website to revalidate cache");
    Ok(())
}

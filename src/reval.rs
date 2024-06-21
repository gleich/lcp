use std::{env, fmt::Display};

use anyhow::{Context, Result};
use reqwest::Client;
use tracing::info;

const REVALIDATE_TOKEN: &str = "REVALIDATE_ACCESS_TOKEN";

pub enum Service {
    Strava,
    Steam,
}

impl Display for Service {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match *self {
            Self::Steam => write!(f, "steam"),
            Self::Strava => write!(f, "strava"),
        }
    }
}

pub async fn call_for_revalidate(client: &Client, service: Service) -> Result<()> {
    client
        .post(format!("https://beta.mattglei.ch/revalidate/{}", service))
        .bearer_auth(env::var(REVALIDATE_TOKEN).context("getting revalidate token env var failed")?)
        .send()
        .await
        .context("sending request to revalidate website cache failed")?;
    info!("made call to website to revalidate cache for {}", service);
    Ok(())
}

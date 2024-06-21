use std::env;

use anyhow::{Context, Result};
use reqwest::Client;
use tracing::info;

const REVALIDATE_TOKEN: &str = "REVALIDATE_ACCESS_TOKEN";

pub enum Service {
    Strava,
    Steam,
}

impl ToString for Service {
    fn to_string(&self) -> String {
        match self {
            &Self::Steam => String::from("steam"),
            &Self::Strava => String::from("strava"),
        }
    }
}

pub async fn call_for_revalidate(client: &Client, service: Service) -> Result<()> {
    client
        .post(format!(
            "https://beta.mattglei.ch/revalidate/{}",
            service.to_string()
        ))
        .bearer_auth(env::var(REVALIDATE_TOKEN).context("getting revalidate token env var failed")?)
        .send()
        .await
        .context("sending request to revalidate website cache failed")?;
    info!("made call to website to revalidate cache");
    Ok(())
}

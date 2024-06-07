use std::env;

use anyhow::{Context, Result};
use chrono::{DateTime, Utc};
use reqwest::Client;
use serde::{Deserialize, Serialize};

const VERCEL_TOKEN: &str = "VERCEL_ACCESS_TOKEN";

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct MainVercelResponse {
    pub projects: Vec<RawProject>,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct RawProject {
    pub name: String,
    #[serde(rename = "updatedAt")]
    pub updated_at: i64,
    pub framework: Option<String>,
    #[serde(rename = "latestDeployments")]
    pub latest_deployments: Vec<Deployment>,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Deployment {
    #[serde(rename = "readyState")]
    pub ready_state: String,
    #[serde(rename = "buildingAt")]
    pub building_at: i64,
    #[serde(rename = "readyAt")]
    pub ready_at: i64,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Project {
    pub name: String,
    pub updated_at: DateTime<Utc>,
    pub framework: Option<String>,
    pub ready_state: String,
    pub time_building: i64,
}

pub async fn fetch_recent(client: &Client) -> Result<Vec<Project>> {
    let mut resp: MainVercelResponse = client
        .get("https://api.vercel.com/v9/projects")
        .bearer_auth(env::var(VERCEL_TOKEN).context("fetching vercel token env var failed")?)
        .send()
        .await
        .context("sending request for recent projects failed")?
        .json()
        .await
        .context("reading json failed from request to get recent projects")?;
    Ok(resp
        .projects
        .iter_mut()
        .map(|p| {
            let deployment = p
                .latest_deployments
                .get_mut(0)
                .context("getting latest deployment failed")
                .unwrap();
            Project {
                name: p.name.to_owned(),
                updated_at: DateTime::from_timestamp_millis(p.updated_at)
                    .context("converting timestamp to datetime")
                    .unwrap(),
                framework: p.framework.to_owned(),
                ready_state: deployment.ready_state.to_owned(),
                time_building: deployment.ready_at - deployment.building_at,
            }
        })
        .collect())
}

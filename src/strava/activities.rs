use anyhow::{Context, Result};
use chrono::{DateTime, Utc};
use reqwest::Client;
use serde::{Deserialize, Serialize};

use crate::strava::token::TokenData;

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Activity {
    pub name: String,
    pub sport_type: String,
    pub start_date: DateTime<Utc>,
    pub map: Map,
    pub trainer: bool,
    pub commute: bool,
    pub private: bool,
    pub average_speed: f32,
    pub max_speed: f32,
    #[serde(default)]
    pub average_temp: i32,
    #[serde(default)]
    pub average_cadence: f32,
    #[serde(default)]
    pub average_watts: f32,
    #[serde(default)]
    pub device_watts: bool,
    #[serde(default)]
    pub average_heartrate: f32,
    pub total_elevation_gain: f32,
    pub moving_time: u32,
    #[serde(default)]
    pub suffer_score: f32,
    pub pr_count: u32,
    pub distance: f32,
    pub id: u64,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Map {
    pub summary_polyline: String,
}

pub async fn fetch_recent(token_data: &TokenData, client: &Client) -> Result<Vec<Activity>> {
    let resp: Vec<Activity> = client
        .get("https://www.strava.com/api/v3/athlete/activities")
        .query(&[("access_token", &token_data.access_token)])
        .send()
        .await
        .context("sending request for recent activities failed")?
        .json()
        .await
        .context("reading json failed from request to get recent activities")?;
    Ok(resp)
}

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
    #[serde(default)]
    pub laps: Option<Vec<Lap>>,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Map {
    pub summary_polyline: String,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Lap {
    pub moving_time: u32,
    pub distance: f32,
    #[serde(default)]
    pub average_watts: f32,
    pub lap_index: u32,
}

pub async fn fetch_recent(token_data: &TokenData, client: &Client) -> Result<Vec<Activity>> {
    let mut resp: Vec<Activity> = client
        .get("https://www.strava.com/api/v3/athlete/activities")
        .query(&[("access_token", &token_data.access_token)])
        .send()
        .await
        .context("sending request for recent activities failed")?
        .json()
        .await
        .context("reading json failed from request to get recent activities")?;
    for activity in &mut resp {
        let laps_resp: Vec<Lap> = client
            .get(format!(
                "https://www.strava.com/api/v3/activities/{}/laps",
                activity.id
            ))
            .query(&[("access_token", &token_data.access_token)])
            .send()
            .await
            .context(format!(
                "sending request for laps for activity of ID {} failed",
                activity.id
            ))?
            .json()
            .await
            .context(format!(
                "reading json failed from request to get laps for activity of ID {}",
                activity.id
            ))?;
        if laps_resp.len() > 1 {
            activity.laps = Some(laps_resp);
        }
    }
    Ok(resp)
}

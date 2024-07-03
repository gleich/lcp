use std::{collections::HashMap, env};

use anyhow::{Context, Result};
use reqwest::Client;
use rocket::{http::Status, post, serde::json::Json};
use serde::Deserialize;
use tracing::info;

use crate::{
    reval::{self, Service},
    strava::{activities, cache, token::TokenData},
};

use super::map;

#[derive(Debug, PartialEq, Deserialize)]
pub struct Event {
    pub aspect_type: String,
    pub event_time: u64,
    pub object_id: u64,
    pub object_type: String,
    pub owner_id: u64,
    pub subscription_id: u64,
    pub updates: HashMap<String, String>,
}

#[post("/", data = "<event>")]
pub async fn endpoint(event: Json<Event>) -> Status {
    dbg!(&event);
    let event_sub_id = event.subscription_id.to_string();
    let expected_sub_id = env::var("STRAVA_SUBSCRIPTION_ID").unwrap();
    if event_sub_id != expected_sub_id {
        info!("subscription_id of {} did not match {}; returning forbidden to request that called endpoint", event_sub_id, expected_sub_id);
        return Status::Forbidden;
    }
    let client = Client::new();
    update(&client)
        .await
        .expect("updating list of activities failed");
    Status::Ok
}

pub async fn update(client: &Client) -> Result<()> {
    let mut token_data = TokenData::new().context("getting strava token data failed")?;
    token_data
        .fetch_if_expired(client)
        .await
        .context("fetching new strava token data if expired failed")?;
    let recent_activities = activities::fetch_recent(&token_data, client)
        .await
        .context("fetching strava recent activities failed")?;
    let updated_cache = cache::update(recent_activities.clone())
        .await
        .expect("updating strava cache failed");
    if updated_cache {
        // generate mapbox images and upload them to S3
        let s3_config = aws_config::load_from_env().await;
        let s3_client = aws_sdk_s3::Client::new(&s3_config);
        map::clear_mapbox_folder(&s3_client)
            .await
            .context("clearing out mapbox folder filled with old maps failed")?;
        for activity in &recent_activities {
            let map = map::fetch_from_mapbox(client, &activity.map.summary_polyline)
                .await
                .context("fetching map from mapbox failed")?;
            map::upload_to_s3(&s3_client, map, activity.id)
                .await
                .context("uploading map to S3 failed")?;
        }
        info!("uploaded {} mapbox images to S3", recent_activities.len());

        reval::call_for_revalidate(client, Service::Strava)
            .await
            .context("calling for website revalidation failed")?;
    }
    Ok(())
}

use std::sync::{Arc, Mutex, MutexGuard, PoisonError};

use chrono::Utc;
use lazy_static::lazy_static;
use rocket::get;
use rocket::serde::json::Json;
use tracing::info;

use crate::resp::Response;
use crate::strava::activities::Activity;
use crate::{auth, metrics};

lazy_static! {
    static ref ACTIVITIES: Arc<Mutex<Response<Vec<Activity>>>> =
        Arc::new(Mutex::new(Response::new(vec![])));
}

#[get("/cache")]
pub fn endpoint(_token: auth::Token) -> Json<Response<Vec<Activity>>> {
    let arc_ref = Arc::clone(&ACTIVITIES);
    let recent_activities = arc_ref.lock().unwrap();
    metrics::REQUEST_SUCCESSFUL_COUNTER.inc();
    metrics::STRAVA_CACHE_REQUEST_COUNTER.inc();
    Json(recent_activities.clone())
}

pub async fn update<'a>(
    recent_activities: Vec<Activity>,
) -> Result<bool, PoisonError<MutexGuard<'a, Response<Vec<Activity>>>>> {
    let mut changer = ACTIVITIES.lock()?;
    if *changer.data != recent_activities {
        changer.data = recent_activities;
        changer.last_updated = Utc::now();
        metrics::STRAVA_CACHE_UPDATE_COUNTER.inc();
        info!("strava cache updated");
        return Ok(true);
    }
    Ok(false)
}

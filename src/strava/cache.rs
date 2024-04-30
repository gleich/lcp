use std::sync::{Arc, Mutex, MutexGuard, PoisonError};

use lazy_static::lazy_static;
use rocket::get;
use rocket::serde::json::Json;
use tracing::info;

use crate::auth;
use crate::strava::activities::Activity;

lazy_static! {
    static ref ACTIVITIES: Arc<Mutex<Vec<Activity>>> = Arc::new(Mutex::new(vec![]));
}

#[get("/cache")]
pub fn endpoint(_token: auth::Token) -> Json<Vec<Activity>> {
    let arc_ref = Arc::clone(&ACTIVITIES);
    let recent_activities = arc_ref.lock().unwrap();
    info!("strava cache endpoint hit");
    Json((recent_activities.clone()).to_vec())
}

pub fn update<'a>(
    recent_activities: Vec<Activity>,
) -> Result<(), PoisonError<MutexGuard<'a, Vec<Activity>>>> {
    let mut changer = ACTIVITIES.lock()?;
    *changer = recent_activities;
    Ok(())
}

use std::sync::{Arc, Mutex, MutexGuard, PoisonError};

use chrono::Utc;
use lazy_static::lazy_static;
use rocket::{get, serde::json::Json};
use tracing::info;

use crate::{auth, github::repos::Repository, metrics, resp::Response};

lazy_static! {
    static ref PINNED_REPOS: Arc<Mutex<Response<Vec<Repository>>>> =
        Arc::new(Mutex::new(Response::new(vec![])));
}

#[get("/cache")]
pub fn endpoint(_token: auth::Token) -> Json<Response<Vec<Repository>>> {
    let arc_ref = Arc::clone(&PINNED_REPOS);
    let pinned_repos = arc_ref.lock().unwrap();
    metrics::SUCCESSFUL_REQUEST_COUNTER.inc();
    metrics::GITHUB_CACHE_REQUEST_COUNTER.inc();
    Json(pinned_repos.clone())
}

pub fn update<'a>(
    pinned_repos: Vec<Repository>,
) -> Result<bool, PoisonError<MutexGuard<'a, Response<Vec<Repository>>>>> {
    let mut changer = PINNED_REPOS.lock()?;
    if *changer.data != pinned_repos {
        changer.data = pinned_repos;
        changer.last_updated = Utc::now();
        metrics::GITHUB_CACHE_UPDATE_COUNTER.inc();
        info!("github cache updated");
        return Ok(true);
    }
    Ok(false)
}

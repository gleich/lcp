use std::sync::{Arc, Mutex, MutexGuard, PoisonError};

use chrono::Utc;
use lazy_static::lazy_static;
use rocket::{get, serde::json::Json};
use tracing::info;

use crate::{auth, metrics, resp::Response};

use super::games::Game;

lazy_static! {
    static ref GAMES: Arc<Mutex<Response<Vec<Game>>>> = Arc::new(Mutex::new(Response::new(vec![])));
}

#[get("/cache")]
pub fn endpoint(_token: auth::Token) -> Json<Response<Vec<Game>>> {
    let arc_ref = Arc::clone(&GAMES);
    let recent_games = arc_ref.lock().unwrap();
    metrics::REQUEST_SUCCESSFUL_COUNT.inc();
    metrics::STEAM_CACHE_REQUEST_COUNT.inc();
    Json(recent_games.clone())
}

pub fn update<'a>(
    recent_games: Vec<Game>,
) -> Result<bool, PoisonError<MutexGuard<'a, Response<Vec<Game>>>>> {
    let mut changer = GAMES.lock()?;
    if *changer.data != recent_games {
        changer.data = recent_games;
        changer.last_updated = Utc::now();
        metrics::STEAM_CACHE_UPDATE_COUNT.inc();
        info!("steam cache updated");
        return Ok(true);
    }
    Ok(false)
}

use std::sync::{Arc, Mutex, MutexGuard, PoisonError};

use chrono::Utc;
use lazy_static::lazy_static;
use rocket::{get, serde::json::Json};
use tracing::info;

use crate::{auth, resp::Response};

use super::games::Game;

lazy_static! {
    static ref GAMES: Arc<Mutex<Response<Vec<Game>>>> = Arc::new(Mutex::new(Response::new(vec![])));
}

#[get("/cache")]
pub fn endpoint(_token: auth::Token) -> Json<Response<Vec<Game>>> {
    let arc_ref = Arc::clone(&GAMES);
    let recent_games = arc_ref.lock().unwrap();
    info!("steam cache endpoint hit");
    Json(recent_games.clone())
}

pub fn update<'a>(
    recent_games: Vec<Game>,
) -> Result<(), PoisonError<MutexGuard<'a, Response<Vec<Game>>>>> {
    let mut changer = GAMES.lock()?;
    if *changer.data != recent_games {
        changer.data = recent_games;
        changer.last_updated = Utc::now();
        info!("steam cache updated")
    }
    Ok(())
}

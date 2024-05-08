use std::sync::{Arc, Mutex, MutexGuard, PoisonError};

use lazy_static::lazy_static;
use rocket::{get, serde::json::Json};
use tracing::info;

use crate::{auth, steam::games::Game};

lazy_static! {
    static ref GAMES: Arc<Mutex<Vec<Game>>> = Arc::new(Mutex::new(vec![]));
}

#[get("/cache")]
pub fn endpoint(_token: auth::Token) -> Json<Vec<Game>> {
    let arc_ref = Arc::clone(&GAMES);
    let recent_games = arc_ref.lock().unwrap();
    info!("steam cache endpoint hit");
    Json((recent_games.clone()).to_vec())
}

pub fn update<'a>(recent_games: Vec<Game>) -> Result<(), PoisonError<MutexGuard<'a, Vec<Game>>>> {
    let mut changer = GAMES.lock()?;
    *changer = recent_games;
    Ok(())
}

use std::env;

use rocket::{
    get,
    serde::{json::Json, Serialize},
    FromForm,
};
use tracing::info;

#[derive(Debug, PartialEq, FromForm)]
pub struct Parameters<'r> {
    pub mode: &'r str,
    pub verify_token: &'r str,
    pub challenge: &'r str,
}

#[derive(Debug, Serialize)]
pub struct Response<'r> {
    #[serde(rename(serialize = "hub.challenge"))]
    pub challenge: &'r str,
}

#[get("/?<hub>")]
pub fn endpoint(hub: Parameters) -> Json<Option<Response>> {
    let verify_token = env::var("STRAVA_VERIFY_TOKEN").unwrap_or_default();
    if hub.verify_token != verify_token {
        info!("received INVALID verify token of {}", hub.verify_token);
        return Json(None);
    }
    info!("received valid verify token of {}", hub.verify_token);
    Json(Some(Response {
        challenge: hub.challenge,
    }))
}

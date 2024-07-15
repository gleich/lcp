use lazy_static::lazy_static;
use prometheus::{register_int_counter, Encoder, IntCounter, TextEncoder};
use rocket::get;

lazy_static! {
    pub static ref REQUEST_COUNT: IntCounter =
        register_int_counter!("request_count", "Number of API requests").unwrap();
    pub static ref REQUEST_SUCCESSFUL_COUNT: IntCounter = register_int_counter!(
        "successful_request_count",
        "Number of successful API requests"
    )
    .unwrap();

    // CACHE SPECIFIC COUNTERS
    pub static ref STRAVA_CACHE_UPDATE_COUNT: IntCounter = register_int_counter!(
        "strava_cache_update_count",
        "Number of updates to the Strava Cache"
    )
    .unwrap();
    pub static ref STRAVA_CACHE_REQUEST_COUNT: IntCounter = register_int_counter!(
        "strava_cache_request_count",
        "Number of valid requests to the Strava Cache"
    )
    .unwrap();
    pub static ref STEAM_CACHE_UPDATE_COUNT: IntCounter = register_int_counter!(
        "steam_cache_update_count",
        "Number of updates to the Steam Cache"
    )
    .unwrap();
    pub static ref STEAM_CACHE_REQUEST_COUNT: IntCounter = register_int_counter!(
        "steam_cache_request_count",
        "Number of valid requests to the Steam Cache"
    )
    .unwrap();
    pub static ref GITHUB_CACHE_UPDATE_COUNT: IntCounter = register_int_counter!(
        "github_cache_update_count",
        "Number of updates to the GitHub Cache"
    )
    .unwrap();
    pub static ref GITHUB_CACHE_REQUEST_COUNT: IntCounter = register_int_counter!(
        "github_cache_request_count",
        "Number of valid requests to the GitHub Cache"
    )
    .unwrap();
}

#[get("/metrics")]
pub fn endpoint() -> String {
    REQUEST_COUNT.inc();
    REQUEST_SUCCESSFUL_COUNT.inc();

    let encoder = TextEncoder::new();
    let mut buffer = vec![];
    let metric_families = prometheus::gather();
    encoder
        .encode(&metric_families, &mut buffer)
        .expect("Failed to encode prometheus data");
    String::from_utf8(buffer).expect("Failed to get families data from buffer")
}

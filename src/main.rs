use std::thread;

use dotenv::dotenv;
use reqwest::Client;
use rocket::{routes, tokio, Config};
use tracing::{info, Level};
use tracing_subscriber::FmtSubscriber;

mod auth;
mod steam;
mod strava;

#[rocket::main]
async fn main() {
    tokio::spawn(async {
        initialize_caches().await;
    });
    let subscriber = FmtSubscriber::builder()
        .with_max_level(Level::INFO)
        .finish();
    tracing::subscriber::set_global_default(subscriber).expect("setting default subscriber failed");
    info!("booted");

    dotenv().expect("Failed to load dotenv");
    initialize_caches().await;

    thread::spawn(steam::update::periodic_update);

    let mut rocket_config = rocket::custom(Config::figment().merge(("address", "0.0.0.0")));
    rocket_config = rocket_config.mount(
        "/strava",
        routes![
            strava::event::endpoint,
            strava::challenge::endpoint,
            strava::cache::endpoint
        ],
    );
    rocket_config
        .mount("/steam", routes![steam::cache::endpoint])
        .launch()
        .await
        .expect("failed to launch rocket");
}

async fn initialize_caches() {
    let client = Client::new();
    strava::event::update(&client)
        .await
        .expect("failed to do initial cache on strava");
    steam::update::cache(&client)
        .await
        .expect("failed to do initial cache of steam");
    info!("initialized caches")
}

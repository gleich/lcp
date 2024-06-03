use std::thread;

use anyhow::{Context, Result};
use dotenv::dotenv;
use reqwest::Client;
use rocket::{launch, routes, Config};
use tracing::{info, Level};
use tracing_subscriber::FmtSubscriber;

mod auth;
mod steam;
mod strava;

#[launch]
async fn rocket() -> _ {
    let subscriber = FmtSubscriber::builder()
        .with_max_level(Level::INFO)
        .finish();
    tracing::subscriber::set_global_default(subscriber).expect("setting default subscriber failed");
    info!("booted");

    dotenv().expect("Failed to load dotenv");
    initialize_caches()
        .await
        .expect("initializing caches failed");

    let mut rocket_config = rocket::custom(Config::figment().merge(("address", "0.0.0.0")));
    rocket_config = rocket_config.mount(
        "/strava",
        routes![
            strava::event::endpoint,
            strava::challenge::endpoint,
            strava::cache::endpoint
        ],
    );
    rocket_config = rocket_config.mount("/steam", routes![steam::cache::endpoint]);
    thread::spawn(|| steam::update::periodic_update());
    rocket_config
}

async fn initialize_caches() -> Result<()> {
    let client = Client::new();
    strava::event::update(&client)
        .await
        .context("failed to do initial cache on strava")?;
    steam::update::cache(&client)
        .await
        .context("failed to do initial cache of steam")?;
    Ok(())
}

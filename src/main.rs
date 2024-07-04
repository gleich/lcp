use std::env;

use anyhow::{Context, Result};
use dotenv::dotenv;
use reqwest::Client;
use rocket::{routes, tokio, Config};
use tracing::{info, Level};
use tracing_subscriber::FmtSubscriber;

mod auth;
mod github;
mod health;
mod resp;
mod steam;
mod strava;

#[rocket::main]
async fn main() {
    // setup logging and reports
    let _guard = sentry::init((
        env::var("SENTRY_URL").expect("getting sentry URL failed"),
        sentry::ClientOptions {
            release: sentry::release_name!(),
            ..Default::default()
        },
    ));
    let subscriber = FmtSubscriber::builder()
        .with_max_level(Level::INFO)
        .finish();
    tracing::subscriber::set_global_default(subscriber).expect("setting default subscriber failed");
    info!("booted");

    dotenv().expect("setting env vars from .env failed");
    initialize_caches().await.expect("setting up caches failed");

    tokio::spawn(async {
        steam::update::periodic_update()
            .await
            .expect("periodic update to steam cache failed");
    });
    tokio::spawn(async {
        github::update::periodic_update()
            .await
            .expect("periodic update to github cache failed");
    });

    let mut rocket_config = rocket::custom(Config::figment().merge(("address", "0.0.0.0")));
    rocket_config = rocket_config.mount("/", routes![health::endpoint]);
    rocket_config = rocket_config.mount(
        "/strava",
        routes![
            strava::event::endpoint,
            strava::challenge::endpoint,
            strava::cache::endpoint
        ],
    );
    rocket_config = rocket_config.mount("/steam", routes![steam::cache::endpoint]);
    rocket_config = rocket_config.mount("/github", routes![github::cache::endpoint]);

    rocket_config
        .launch()
        .await
        .expect("failed to launch rocket");
}

async fn initialize_caches() -> Result<()> {
    let client = Client::new();
    strava::event::update(&client)
        .await
        .context("failed to do initial cache on strava")?;
    steam::update::cache(&client)
        .await
        .context("failed to do initial cache of steam")?;
    github::update::cache(&client)
        .await
        .context("failed to do initial cache on github")?;
    info!("initialized caches");
    Ok(())
}

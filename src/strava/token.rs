use std::{
    env,
    time::{SystemTime, UNIX_EPOCH},
};

use anyhow::{Context, Result};
use reqwest::Client;
use serde::Deserialize;

#[derive(Debug, Deserialize)]
pub struct TokenData {
    pub access_token: String,
    pub refresh_token: String,
    pub expires_at: u32,
}

const ACCESS_TOKEN: &str = "STRAVA_ACCESS_TOKEN";
const REFRESH_TOKEN: &str = "STRAVA_REFRESH_TOKEN";
const REFRESH_TOKEN_EXPIRATION: &str = "STRAVA_REFRESH_TOKEN_EXPIRATION";

impl TokenData {
    pub fn new() -> Result<Self> {
        Ok(Self {
            access_token: env::var(ACCESS_TOKEN).context("fetching access token env var failed")?,
            refresh_token: env::var(REFRESH_TOKEN)
                .context("fetching refresh token env var failed")?,
            expires_at: env::var(REFRESH_TOKEN_EXPIRATION)
                .context("fetching refresh token expiration env var failed")?
                .parse()?,
        })
    }

    pub async fn fetch_if_expired(&mut self, client: &Client) -> Result<()> {
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .context("getting current unix time failed")?
            .as_secs() as u32;
        // if token hasn't expired yet, don't do anything
        if self.expires_at > now {
            return Ok(());
        }

        let resp: TokenData = client
            .post("https://www.strava.com/oauth/token")
            .query(&[
                ("client_id", &env::var("STRAVA_CLIENT_ID").unwrap()),
                ("client_secret", &env::var("STRAVA_CLIENT_SECRET").unwrap()),
                ("grant_type", &String::from("refresh_token")),
                ("refresh_token", &self.refresh_token),
                ("code", &env::var("STRAVA_OAUTH_CODE").unwrap()),
            ])
            .send()
            .await
            .context("sending request for new refresh token failed")?
            .json()
            .await
            .context("reading json failed from request to get refresh token")?;

        env::set_var(ACCESS_TOKEN, &resp.access_token);
        env::set_var(REFRESH_TOKEN, &resp.refresh_token);
        env::set_var(REFRESH_TOKEN_EXPIRATION, resp.expires_at.to_string());

        self.access_token = resp.access_token;
        self.refresh_token = resp.refresh_token;
        self.expires_at = resp.expires_at;

        dbg!(self);

        Ok(())
    }
}

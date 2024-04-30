use std::env;

use rocket::{
    http::Status,
    request::FromRequest,
    request::{Outcome, Request},
};

pub struct Token(());

// https://github.com/cjdenio/shorty/blob/master/src/auth.rs
#[rocket::async_trait]
impl<'r> FromRequest<'r> for Token {
    type Error = String;

    async fn from_request(req: &'r Request<'_>) -> Outcome<Self, Self::Error> {
        let auth_header = match req.headers().get_one("Authorization") {
            Some(x) => x,
            None => {
                return Outcome::Error((Status::Unauthorized, String::from("Invalid API token.")))
            }
        };

        let provided_token = match auth_header.strip_prefix("Bearer ") {
            Some(x) => x,
            None => {
                return Outcome::Error((Status::Unauthorized, String::from("Invalid API token.")))
            }
        };

        let valid_tokens = match env::var("VALID_TOKENS") {
            Ok(x) => x,
            Err(_) => {
                return Outcome::Error((Status::Unauthorized, String::from("Invalid API token.")))
            }
        };

        if !valid_tokens
            .split(',')
            .collect::<Vec<&str>>()
            .contains(&provided_token)
        {
            return Outcome::Error((Status::Unauthorized, String::from("Invalid API token.")));
        }

        Outcome::Success(Token(()))
    }
}

use std::env;

use anyhow::{Context, Result};
use chrono::{DateTime, Utc};
use reqwest::{header::USER_AGENT, Client};
use serde::{Deserialize, Serialize};

const GITHUB_TOKEN: &str = "GITHUB_ACCESS_TOKEN";

pub mod raw {
    use chrono::{DateTime, Utc};
    use serde::{Deserialize, Serialize};

    #[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
    pub struct Response {
        pub data: Data,
    }

    #[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
    pub struct Data {
        pub viewer: Viewer,
    }

    #[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
    pub struct Viewer {
        #[serde(rename = "pinnedItems")]
        pub pinned_items: PinnedItems,
    }

    #[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
    pub struct PinnedItems {
        pub nodes: Vec<Repository>,
    }

    #[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
    pub struct Repository {
        pub name: String,
        pub owner: Owner,
        #[serde(rename = "primaryLanguage")]
        pub primary_language: PrimaryLanguage,
        pub description: String,
        #[serde(rename = "updatedAt")]
        pub updated_at: DateTime<Utc>,
        #[serde(rename = "stargazerCount")]
        pub stargazer_count: u32,
    }

    #[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
    pub struct Owner {
        pub login: String,
    }

    #[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
    pub struct PrimaryLanguage {
        pub name: String,
        pub color: String,
    }
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Repository {
    pub name: String,
    pub owner: String,
    pub language: String,
    pub language_color: String,
    pub description: String,
    pub updated_at: DateTime<Utc>,
    pub stargazers: u32,
}

pub async fn fetch_pinned_repos(client: &Client) -> Result<Vec<Repository>> {
    let resp = client
        .post("https://api.github.com/graphql")
        .bearer_auth(env::var(GITHUB_TOKEN).context("loading GITHUB access token failed")?)
        .header(USER_AGENT, "lcp/1.0")
        .body(r#"{"query": "query{viewer{pinnedItems(first:6,types:REPOSITORY){nodes{... on Repository{name owner{login}primaryLanguage{name color}description updatedAt stargazerCount isPrivate}}}}}","variables": {}}"#,
        )
        .send()
        .await
        .context("sending request to GraphQL API for recent games failed")?;
    let resp_text = resp
        .text()
        .await
        .context("getting raw response text failed")?;
    let mut raw_data: raw::Response = serde_json::from_str(&resp_text)
        .context(format!(
            "reading json failed from request to get data from graphql api: {}",
            resp_text
        ))
        .context("reading json failed from request to get pinned repos")?;
    Ok(raw_data
        .data
        .viewer
        .pinned_items
        .nodes
        .iter_mut()
        .map(|a| Repository {
            name: a.name.to_owned(),
            owner: a.owner.login.to_owned(),
            language: a.primary_language.name.to_owned(),
            language_color: a.primary_language.color.to_owned(),
            description: a.description.to_owned(),
            updated_at: a.updated_at,
            stargazers: a.stargazer_count,
        })
        .collect())
}

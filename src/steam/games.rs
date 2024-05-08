use serde::{Deserialize, Serialize};

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Game {
    pub name: String,
    pub img_icon_url: String,
    pub playtime: Playtime,
}

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Playtime {
    pub forever: u32,
    pub last_2week: u32,
}

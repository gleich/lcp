use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

#[derive(Debug, PartialEq, Deserialize, Serialize, Clone)]
pub struct Response<T> {
    pub last_updated: DateTime<Utc>,
    pub data: T,
}

impl<T> Response<T> {
    pub fn new(t: T) -> Response<T> {
        Response {
            last_updated: Utc::now(),
            data: t,
        }
    }
}

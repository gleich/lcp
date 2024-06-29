use std::env;

use anyhow::{Context, Result};
use aws_sdk_s3::primitives::ByteStream;
use rocket::http::hyper::body::Bytes;
use tracing::info;

const MAPBOX_TOKEN: &str = "MAPBOX_ACCESS_TOKEN";

const BUCKET_NAME: &str = "gleich";
const S3_FOLDER_NAME: &str = "mapbox-maps/";

pub async fn fetch_from_mapbox(client: &reqwest::Client, polyline: &str) -> Result<Bytes> {
    let line_width = 2.0;
    let line_color = "000";
    let size = 230;
    let resp = client
        .get(format!(
        "https://api.mapbox.com/styles/v1/mattgleich/clxxsfdfm002401qj7jcxh47e/static/path-{}+{}({})/auto/{}x{}@2x",
        line_width, line_color, urlencoding::encode(polyline), size, size
    ))
        .query(&[(
            "access_token",
            &env::var(MAPBOX_TOKEN).context("getting mapbox access token from env vars failed")?,
        )])
        .send()
        .await
        .context("sending request for static mapbox map image failed")?
        .bytes()
        .await
        .context("decoding reqwest for static mapbox map image failed")?;
    Ok(resp)
}

pub async fn clear_mapbox_folder(client: &aws_sdk_s3::Client) -> Result<()> {
    let imgs = client
        .list_objects_v2()
        .bucket(BUCKET_NAME)
        .send()
        .await
        .context("listing out images currently stored in S3 failed")?;
    for object in imgs.contents() {
        let key = object.key().unwrap().to_string();
        if key.starts_with(S3_FOLDER_NAME) {
            client
                .delete_object()
                .bucket(BUCKET_NAME)
                .key(key)
                .send()
                .await?;
        }
    }
    info!("reset mapbox S3 bucket");
    Ok(())
}

pub async fn upload_to_s3(client: &aws_sdk_s3::Client, image_data: Bytes, id: u64) -> Result<()> {
    client
        .put_object()
        .bucket(BUCKET_NAME)
        .key(format!("{}{}.png", S3_FOLDER_NAME, id))
        .body(ByteStream::from(image_data))
        .send()
        .await?;
    Ok(())
}

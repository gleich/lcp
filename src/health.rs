use rocket::get;

#[get("/health")]
pub fn endpoint() -> String {
    String::from("OK")
}

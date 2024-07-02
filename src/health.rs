use rocket::get;

#[get("/health")]
pub fn endpoint() -> String {
    return String::from("OK");
}

pub struct AuthService;

impl AuthService {
    pub fn is_valid_token(token: &str) -> bool {
        let password = std::env::var("MOTE_PASSWORD").expect("Password is not set");
        token == password
    }
}

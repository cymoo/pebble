use crate::errors::{bad_request, ApiError, ApiResult};
use crate::service::auth_service::AuthService;
use axum::extract::Request;
use axum::http::header;
use axum::middleware::Next;
use axum::response::Response;

/// Middleware function to validate tokens in incoming requests.
///
/// This function checks if the request path is in the list of paths that skip token verification (`skip_paths`).
/// If the path requires verification, it extracts the token from the `Cookie` or `Authorization` header,
/// and checks if the token is valid using the `AuthService::is_valid_token` method.
///
/// # Arguments
/// * `skip_paths` - A list of paths that should skip token verification.
/// * `request` - The incoming HTTP request.
/// * `next` - The next middleware or handler in the chain.
///
/// # Returns
/// * `AppResult<Response>` - Returns the response from the next middleware/handler if the token is valid or the path is skipped.
///   Otherwise, returns an error indicating the reason for failure (e.g., missing or unauthorized token).
pub async fn check_access(skip_paths: &[&str], request: Request, next: Next) -> ApiResult<Response> {
    let path = request.uri().path();

    // Check if verification should be skipped
    if skip_paths.iter().any(|skip_path| path.starts_with(skip_path)) {
        return Ok(next.run(request).await);
    }

    let token = get_cookie(&request, "token")
        .or(extract_bearer(&request))
        .ok_or(bad_request("No token provided"))?;

    if !AuthService::is_valid_token(&token) {
        return Err(ApiError::Unauthorized("Invalid token".to_string()));
    }

    let response = next.run(request).await;
    Ok(response)
}

fn extract_bearer(request: &Request) -> Option<String> {
    let auth_header = request.headers().get(header::AUTHORIZATION)?;
    let auth_str = auth_header.to_str().ok()?;
    let token = auth_str.strip_prefix("Bearer ")?;

    Some(token.to_string())
}

fn get_cookie(request: &Request, name: &str) -> Option<String> {
    let cookie_header = request.headers().get(header::COOKIE)?;
    let cookie_str = cookie_header.to_str().ok()?;

    cookie_str
        .split(';')
        .find_map(|s| {
            let mut parts = s.trim().splitn(2, '=');
            let cookie_name = parts.next()?;
            let cookie_value = parts.next()?;

            if cookie_name == name {
                Some(cookie_value.to_string())
            } else {
                None
            }
        })
}

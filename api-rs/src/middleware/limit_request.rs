use crate::config::rd::RedisPool;
use crate::errors::ApiError::TooManyRequests;
use crate::errors::ApiResult;
use anyhow::{Context, Result};
use axum::extract::Request;
use axum::middleware::Next;
use axum::response::Response;
use redis::ExistenceCheck::NX;
use redis::SetExpiry::EX;
use redis::SetOptions;

/// Middleware function to enforce rate limiting for incoming requests.
///
/// This function checks if the number of requests for a specific path (used as the Redis key) has exceeded
/// the allowed limit (`max_count`) within a given time window (`expires`). If the limit is exceeded,
/// a `TooManyRequests` error is returned. Otherwise, the request is passed to the next middleware or handler.
///
/// # Arguments
/// * `pool` - A connection pool to the Redis instance for tracking request counts.
/// * `expires` - The expiration time (in seconds) for the rate limit window.
/// * `max_count` - The maximum number of requests allowed within the time window.
/// * `req` - The incoming HTTP request.
/// * `next` - The next middleware or handler in the chain.
///
/// # Returns
/// * `AppResult<Response>` - Returns the response from the next middleware/handler if the rate limit is not exceeded.
///   If the limit is exceeded, a `TooManyRequests` error is returned.
pub async fn limit_request(
    pool: RedisPool,
    expires: u64,
    max_count: u64,
    req: Request,
    next: Next,
) -> ApiResult<Response> {
    let key = format!("rate:{}", req.uri().path());

    let below_limit = check_rate_limit(&pool, &key, expires, max_count).await?;
    if !below_limit {
        return Err(TooManyRequests("Too many attempts, try again later".to_owned()));
    }

    Ok(next.run(req).await)
}

/// Checks if the rate limit for a given key has been exceeded.
///
/// This function uses Redis to track the number of requests made for a specific key within a given time window.
/// It sets the key with an expiration time if it doesn't already exist, increments the request count, and checks
/// if the count exceeds the allowed maximum (`max_count`).
///
/// # Arguments
/// * `pool` - A connection pool to the Redis instance.
/// * `key` - The Redis key used to track the rate limit (e.g., a user ID or request path).
/// * `expires` - The expiration time (in seconds) for the key, defining the rate limit window.
/// * `max_count` - The maximum number of requests allowed within the time window.
///
/// # Returns
/// * `Result<bool>` - Returns `Ok(true)` if the request count is within the limit, or `Ok(false)` if the limit has been exceeded.
///   If an error occurs (e.g., Redis connection or query failure), an `Err` is returned.
pub async fn check_rate_limit(
    pool: &RedisPool,
    key: &str,
    expires: u64,
    max_count: u64,
) -> Result<bool> {
    let mut conn = pool.get().await?;

    let rv: [u64; 1] = redis::pipe()
        .atomic()
        .set_options(
            &key,
            0,
            SetOptions::default()
                .with_expiration(EX(expires))
                .conditional_set(NX),
        )
        .ignore()
        .incr(&key, 1)
        .query_async(&mut *conn).await.context("Redis Error")?;

    Ok(rv[0] <= max_count)
}

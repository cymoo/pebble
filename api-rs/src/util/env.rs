use anyhow::{anyhow, Context, Result};
use dotenvy::dotenv;
use std::env;
use std::path::Path;
use std::str::FromStr;
use std::sync::OnceLock;

// A static variable to ensure that environment variables are loaded only once.
static LOAD_ENV: OnceLock<()> = OnceLock::new();

/// Loads environment variables from `.env` and environment-specific files.
///
/// This function initializes environment variables by loading them from `.env` files.
/// It follows a specific order of precedence:
/// 1. Loads the default `.env` file.
/// 2. Loads an environment-specific file (`.env.dev` for debug mode or `.env.prod` for production mode).
/// 3. Loads a local override file (`.env.local`) if it exists.
pub fn load_dotenv() {
    LOAD_ENV.get_or_init(|| {
        // load .env
        dotenv().ok();

        let debug = cfg!(debug_assertions);
        let env_file = if debug { ".env.dev" } else { ".env.prod" };

        // load .env.dev or .env.prod
        if Path::new(env_file).exists() {
            dotenvy::from_filename(env_file).ok();
        }

        // load .env.local
        if Path::new(".env.local").exists() {
            dotenvy::from_filename(".env.local").ok();
        }
    });
}

/// Retrieves a value from an environment variable and parses it into type `T`.
/// If the variable is not set, returns `default`. If parsing fails, returns an error.
pub fn get_env_or<T>(key: &str, default: T) -> Result<T>
where
    T: FromStr,
    T::Err: std::fmt::Debug,
{
    match env::var(key) {
        Ok(val) => val.parse()
            .map_err(|_| anyhow!(format!("Failed to parse {} env var", key))),
        Err(_) => Ok(default)
    }
}

/// Retrieves a vector from an environment variable.
/// If the variable is not set, returns `default`. If parsing fails, returns an error.
pub fn get_vec_from_env_or<T>(key: &str, default: Vec<T>) -> Result<Vec<T>>
where
    T: FromStr,
    T::Err: Into<anyhow::Error>,
{
    match env::var(key) {
        Ok(val) => val.split(',')
            .map(|s| s.trim().parse().map_err(Into::into)
                .context(format!("Failed to parse {} env var", key)))
            .collect(),
        Err(_) => Ok(default),
    }
}

/// Retrieves a `u64` from an environment variable.
/// Supporting K, M, G suffixes (case-insensitive).
/// If the variable is not set, returns `default`. If parsing fails, returns an error.
pub fn get_size_from_env_or(key: &str, default: u64) -> Result<u64> {
    match env::var(key) {
        Ok(val) => parse_size(&val)
            .ok_or(anyhow!(format!("Failed to parse {} env var", key))),
        Err(_) => Ok(default),
    }
}

/// Retrieves a `bool` from an environment variable.
/// Recognizes `"true"`, `"1"`, `"yes"`, `"on"` as `true`; `"false"`, `"0"`, `"no"`, `"off"` as `false`.
/// If the variable is not set, returns `default`. If parsing fails, returns an error.
pub fn get_bool_from_env_or(key: &str, default: bool) -> Result<bool> {
    match env::var(key) {
        Ok(value) => {
            let value = value.to_lowercase();
            match value.as_str() {
                "true" | "1" | "yes" | "on" => Ok(true),
                "false" | "0" | "no" | "off" => Ok(false),
                _ => Err(anyhow!(format!("Failed to parse {} env var as `bool`", key)))
            }
        }
        Err(_) => Ok(default),
    }
}

/// Converts a size string to a number, supporting K, M, G suffixes (case-insensitive)
///
/// # Arguments
/// * `size_str` - The size string to parse
///
/// # Returns
/// Parsed numeric size, or None if parsing fails
/// ```
pub fn parse_size(size_str: &str) -> Option<u64> {
    if size_str.is_empty() {
        return None;
    }

    let size_str = size_str.to_lowercase();

    // Split into numeric part and unit multiplier
    let (num_part, unit_multiplier) = match size_str.chars().last() {
        Some('k') => (&size_str[..size_str.len() - 1], 1024u64),
        Some('m') => (&size_str[..size_str.len() - 1], 1024u64 * 1024),
        Some('g') => (&size_str[..size_str.len() - 1], 1024u64 * 1024 * 1024),
        _ => (size_str.as_str(), 1),
    };

    // Try to parse the numeric part and multiply by unit multiplier
    match num_part.parse::<u64>() {
        Ok(num) => Some(num * unit_multiplier),
        Err(_) => None,
    }
}

/// Measures and logs the execution time of a code block.
///
/// Supports both synchronous and asynchronous code. Logs the elapsed time using `info!`.
///
/// # Usage
/// - Sync: `timeit!("Task", { /* code */ })`
/// - Async: `timeit!("Task", async { /* code */ }, async)`
///
#[macro_export]
macro_rules! timeit {
    ($expr:expr) => {
        timeit!("Time elapsed", $expr)
    };

    ($expr:expr, async) => {
        timeit!("Time elapsed", $expr, async)
    };

    ($prefix:expr, $expr:expr) => {{
        let start = std::time::Instant::now();
        let result = $expr;
        let duration = start.elapsed();
        info!("{}: {:?}", $prefix, duration);
        result
    }};

    ($prefix:expr, $expr:expr, async) => {{
        let start = std::time::Instant::now();
        let result = $expr.await;
        let duration = start.elapsed();
        info!("{}: {:?}", $prefix, duration);
        result
    }};
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_size() {
        assert_eq!(parse_size(""), None);
        assert_eq!(parse_size("1"), Some(1));
        assert_eq!(parse_size("100"), Some(100));
        assert_eq!(parse_size("3k"), Some(3 * 1024));
        assert_eq!(parse_size("100M"), Some(100 * 1024 * 1024));
        assert_eq!(parse_size("5g"), Some(5 * 1024 * 1024 * 1024));
        assert_eq!(parse_size("5 g"), None);
        assert_eq!(parse_size("abc"), None);
    }

}

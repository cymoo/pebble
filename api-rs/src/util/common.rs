use anyhow::{anyhow, Context, Result};
use chrono::{DateTime, FixedOffset, NaiveDate, NaiveTime, TimeZone};
use dotenvy::dotenv;
use std::env;
use std::path::Path;
use std::str::FromStr;
use std::sync::OnceLock;

/// The Pipe trait provides a method to pipe a value through a transformation.
///
/// This trait allows for a more functional programming style by enabling
/// method chaining and easy value transformation.
///
/// # Examples
///
/// ```rust
/// use pebble::util::common::Pipe;
/// let result = 5.pipe(|x| x * 2);  // result is 10
/// let string = "hello".pipe(|s| s.to_uppercase());  // string is "HELLO"
/// ```
pub trait Pipe {
    /// Transforms the current value by applying the given function.
    ///
    /// # Arguments
    ///
    /// * `f` - A closure that takes the current value and returns a transformed value
    ///
    /// # Returns
    ///
    /// The result of applying the transformation function to the current value
    fn pipe<F, R>(self, f: F) -> R
    where
        F: FnOnce(Self) -> R,
        Self: Sized;
}

impl<T> Pipe for T {
    fn pipe<F, R>(self, f: F) -> R
    where
        F: FnOnce(Self) -> R,
        Self: Sized,
    {
        // Apply the transformation function to the current value
        f(self)
    }
}

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

/// Replaces the starting substring `from` in `s` with `to` if `s` starts with `from`.
pub fn replace_from_start(s: &str, from: &str, to: &str) -> String {
    if s.starts_with(from) {
        let remainder = &s[from.len()..];
        format!("{}{}", to, remainder)
    } else {
        s.to_string()
    }
}

/// Convert a date string to a DateTime object with timezone information
///
/// # Arguments
///
/// * `date_str` - The date string in "yyyy-MM-dd" format
/// * `offset` - Timezone offset in minutes
/// * `end_of_day` - Whether to use the end time of the day (defaults to false, which means the start time of the day)
///
/// # Errors
///
/// Returns an error if:
/// * The timezone offset is out of range (-1440 to 1440 minutes)
/// * The date string cannot be parsed
/// * The time components cannot be combined
pub fn to_datetime(
    date_str: &str,
    offset: i32,
    end_of_day: bool,
) -> Result<DateTime<FixedOffset>> {
    // Validate timezone offset
    if offset.abs() > 1440 {
        return Err(anyhow::anyhow!(
            "Timezone offset must be between -1440 and 1440 minutes: {offset}"
        ));
    }

    // Parse the date string and create time component
    let local_datetime = NaiveDate::parse_from_str(date_str, "%Y-%m-%d")?.and_time(
        if end_of_day {
            NaiveTime::from_hms_milli_opt(23, 59, 59, 999)
        } else {
            NaiveTime::from_hms_opt(0, 0, 0)
        }
            .expect("Invalid time components"),
    );

    // Create timezone offset and convert local time
    FixedOffset::east_opt(offset * 60)
        .ok_or_else(|| anyhow::anyhow!("Invalid timezone offset: {offset}"))?
        .from_local_datetime(&local_datetime)
        .earliest()
        .ok_or_else(|| anyhow::anyhow!("Invalid datetime conversion"))
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
    use chrono::Timelike;

    #[cfg(test)]
    mod replace_tests {
        use crate::util::common::replace_from_start;

        #[test]
        fn test_basic_replacement() {
            assert_eq!(replace_from_start("foobar", "foo", "new"), "newbar");
            assert_eq!(replace_from_start("apple banana", "apple", "fruit"), "fruit banana");
        }

        #[test]
        fn test_no_replacement() {
            assert_eq!(replace_from_start("foobar", "bar", "new"), "foobar");
            assert_eq!(replace_from_start("test", "longer", "x"), "test");
        }

        #[test]
        fn test_empty_from() {
            assert_eq!(replace_from_start("foobar", "", "prefix"), "prefixfoobar");
            assert_eq!(replace_from_start("", "", "x"), "x");
        }

        #[test]
        fn test_empty_string() {
            assert_eq!(replace_from_start("", "abc", "x"), "");
            assert_eq!(replace_from_start("", "", ""), "");
        }

        #[test]
        fn test_unicode() {
            assert_eq!(replace_from_start("你好啊世界", "你好啊", "美好的"), "美好的世界");
            assert_eq!(replace_from_start("🦀🍰", "🦀", "🐻"), "🐻🍰");
            assert_eq!(replace_from_start("こんにちは", "こん", "おは"), "おはにちは");
        }
    }

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

    #[test]
    fn test_start_of_day() {
        let dt = to_datetime("2024-01-21", 480, false).unwrap();
        assert_eq!(dt.hour(), 0);
        assert_eq!(dt.minute(), 0);
        assert_eq!(dt.second(), 0);
        assert_eq!(dt.nanosecond(), 0);
        assert_eq!(dt.offset().local_minus_utc(), 480 * 60);
    }

    #[test]
    fn test_end_of_day() {
        let dt = to_datetime("2024-01-21", 480, true).unwrap();
        assert_eq!(dt.hour(), 23);
        assert_eq!(dt.minute(), 59);
        assert_eq!(dt.second(), 59);
        assert_eq!(dt.nanosecond(), 999_000_000);
        assert_eq!(dt.offset().local_minus_utc(), 480 * 60);
    }

    #[test]
    fn test_invalid_offset() {
        assert!(to_datetime("2024-01-21", 1441, false).is_err());
        assert!(to_datetime("2024-01-21", -1441, false).is_err());
    }

    #[test]
    fn test_invalid_date() {
        assert!(to_datetime("invalid-date", 480, false).is_err());
    }
}

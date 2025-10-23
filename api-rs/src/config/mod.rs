use crate::util::common::{get_env_or, get_size_from_env_or, get_vec_from_env_or, load_dotenv};
use tower_http::cors::{CorsLayer, AllowHeaders, AllowOrigin, AllowMethods, Any};
use std::time::Duration;
use std::env;
use std::str::FromStr;
use std::fmt::Debug;

pub mod db;
pub mod rd;

#[derive(Debug, Clone)]
pub struct AppConfig {
    // Basic app info
    pub app_name: String,
    pub app_version: String,

    // App settings
    pub posts_per_page: u32,
    pub static_url: String,
    pub static_path: String,

    // Server settings
    pub http: HTTPConfig,
    pub upload: UploadConfig,
    pub db: DBConfig,
    pub redis: RedisConfig,
    pub log: LogConfig,
}

#[derive(Debug, Clone)]
pub struct HTTPConfig {
    pub ip: String,
    pub port: u16,
    pub max_body_size: u64,
    pub read_timeout_secs: u64,
    pub write_timeout_secs: u64,
    pub idle_timeout_secs: u64,
    pub cors: CORSConfig,
}

#[derive(Debug, Clone)]
pub struct UploadConfig {
    pub base_path: String,
    pub base_url: String,
    pub thumb_width: u32,
    pub image_formats: Vec<String>,
}

#[derive(Debug, Clone)]
pub struct DBConfig {
    pub url: String,
    pub pool_size: u32,
    pub auto_migrate: bool,
}

#[derive(Debug, Clone)]
pub struct RedisConfig {
    pub url: String,
    pub password: Option<String>,
    pub db: u32
}

#[derive(Debug, Clone)]
pub struct CORSConfig {
    pub allowed_origins: Vec<String>,
    pub allowed_methods: Vec<String>,
    pub allowed_headers: Vec<String>,
    pub allow_credentials: bool,
    pub max_age: u64,
}

#[derive(Debug, Clone)]
pub struct LogConfig {
    pub log_requests: bool,
}

impl AppConfig {
    pub fn from_env() -> Self {
        load_dotenv();

        let app_name = get_env_or("APP_NAME", "Pebble".to_string()).unwrap();
        let app_version = get_env_or("APP_VERSION", "1.0.0".to_string()).unwrap();

        let posts_per_page = get_env_or("POSTS_PER_PAGE", 30).unwrap();
        let static_url = get_env_or("STATIC_URL", "/static".to_string()).unwrap();
        let static_path = get_env_or("STATIC_PATH", "./static".to_string()).unwrap();

        AppConfig {
            app_name,
            app_version,

            posts_per_page,
            static_url,
            static_path,

            http: HTTPConfig::from_env(),
            upload: UploadConfig::from_env(),
            db: DBConfig::from_env(),
            redis: RedisConfig::from_env(),
            log: LogConfig::from_env(),
        }
    }
}

impl HTTPConfig {
    pub fn from_env() -> Self {
        load_dotenv();

        let ip = get_env_or("HTTP_IP", "127.0.0.1".to_string()).unwrap();
        let port = get_env_or("HTTP_PORT", 8000).unwrap();
        let max_body_size = get_size_from_env_or("HTTP_MAX_BODY_SIZE", 10 * 1024 * 1024).unwrap();
        let read_timeout_secs = get_env_or("HTTP_READ_TIMEOUT_SECS", 10).unwrap();
        let write_timeout_secs = get_env_or("HTTP_WRITE_TIMEOUT_SECS", 10).unwrap();
        let idle_timeout_secs = get_env_or("HTTP_IDLE_TIMEOUT_SECS", 30).unwrap();
        let cors = CORSConfig::from_env();
        HTTPConfig {
            ip,
            port,
            max_body_size,
            read_timeout_secs,
            write_timeout_secs,
            idle_timeout_secs,
            cors,
        }
    }
}

impl UploadConfig {
    pub fn from_env() -> Self {
        load_dotenv();

        let base_path = get_env_or("UPLOAD_PATH", "./uploads".to_string()).unwrap();
        let base_url = get_env_or("UPLOAD_URL", "/uploads".to_string()).unwrap();
        let thumb_width = get_env_or("UPLOAD_THUMB_WIDTH", 128).unwrap();
        let image_formats = get_vec_from_env_or("UPLOAD_IMAGE_FORMATS", strs_to_strings(vec!["jpeg", "jpg", "png", "webp", "gif"])).unwrap();

        UploadConfig {
            base_path,
            base_url,
            thumb_width,
            image_formats,
        }
    }
}

impl DBConfig {
    pub fn from_env() -> Self {
        load_dotenv();

        let url = get_env_or("DATABASE_URL", "app.db".to_string()).unwrap();
        let pool_size = get_env_or("DATABASE_POOL_SIZE", 5).unwrap();
        let auto_migrate = get_env_or("DATABASE_AUTO_MIGRATE", true).unwrap();

        DBConfig {
            url,
            pool_size,
            auto_migrate,
        }
    }
}

impl RedisConfig {
    pub fn from_env() -> Self {
        load_dotenv();

        let url = get_env_or("REDIS_URL", "redis://localhost:6379".to_string()).unwrap();
        let password = env::var("REDIS_PASSWORD").ok();
        let db = get_env_or("REDIS_DB", 0).unwrap();

        RedisConfig {
            url,
            password,
            db,
        }
    }
}

impl CORSConfig {
    pub fn from_env() -> Self {
        load_dotenv();

        let allowed_origins = get_vec_from_env_or("CORS_ALLOWED_ORIGINS", vec![]).unwrap();
        let allowed_methods = get_vec_from_env_or("CORS_ALLOWED_METHODS", strs_to_strings(vec!["GET", "POST", "PUT", "DELETE", "OPTIONS"])).unwrap();
        let allowed_headers = get_vec_from_env_or("CORS_ALLOWED_HEADERS", vec!["Content-Type".to_string(), "Authorization".to_string()]).unwrap();
        let allow_credentials = get_env_or("CORS_ALLOW_CREDENTIALS", false).unwrap();
        let max_age = get_env_or("CORS_MAX_AGE", 86400).unwrap();

        CORSConfig {
            allowed_origins,
            allowed_methods,
            allowed_headers,
            allow_credentials,
            max_age,
        }
    }

    pub fn into_layer(self) -> CorsLayer {
        let mut cors = CorsLayer::new();

        cors = if self.allowed_origins.contains(&"*".to_string()) {
            cors.allow_origin(Any)
        } else {
            cors.allow_origin(AllowOrigin::list(convert_vec( self.allowed_origins.clone())))
        };

        cors = if self.allowed_methods.contains(&"*".to_string()) {
            cors.allow_methods(Any)
        } else {
            cors.allow_methods(AllowMethods::list(convert_vec( self.allowed_methods.clone())))
        };

        cors = if self.allowed_headers.contains(&"*".to_string()) {
            cors.allow_headers(Any)
        } else {
            cors.allow_headers(AllowHeaders::list(convert_vec( self.allowed_headers.clone())))
        };

        cors = cors
            .allow_credentials(self.allow_credentials)
            .max_age(Duration::from_secs(self.max_age));

        cors
    }
}

impl LogConfig {
    pub fn from_env() -> Self {
        load_dotenv();

        let log_requests = get_env_or("LOG_REQUESTS", false).unwrap();

        LogConfig {
            log_requests,
        }
    }
}

// convert vectors of &str to owned Strings
fn strs_to_strings(vec: Vec<&str>) -> Vec<String> {
    vec.into_iter().map(|s| s.to_string()).collect()
}

// Helper function to convert Vec<String> to Vec<T>
fn convert_vec<T: FromStr>(strings: Vec<String>) -> Vec<T>
where
    <T as FromStr>::Err: Debug,
{
    strings.into_iter().map(|s| s.parse().unwrap()).collect()
}

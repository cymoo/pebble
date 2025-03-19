use crate::util::common::{get_bool_from_env_or, get_env_or, get_size_from_env_or, get_vec_from_env_or, load_dotenv};
use std::env;

pub mod db;
pub mod rd;

#[derive(Debug, Clone)]
pub struct UploadConfig {
    pub upload_dir: String,
    pub upload_url: String,
    pub thumbnail_size: u32,
    pub image_formats: Vec<String>,
}

impl UploadConfig {
    pub fn from_env() -> Self {
        load_dotenv();

        let upload_dir = get_env_or("UPLOAD_DIR", "./uploads".to_string()).unwrap();
        let upload_url = get_env_or("UPLOAD_URL", "/uploads".to_string()).unwrap();
        let thumbnail_size = get_env_or("THUMBNAIL_SIZE", 128).unwrap();
        let image_formats = get_vec_from_env_or("IMAGE_FORMATS", vec![]).unwrap();

        UploadConfig {
            upload_dir,
            upload_url,
            thumbnail_size,
            image_formats,
        }
    }
}

#[derive(Debug, Clone)]
pub struct SearchConfig {
    pub partial_match: bool,
    pub max_results: usize,
    pub key_prefix: String,
}

impl SearchConfig {
    pub fn from_env() -> Self {
        load_dotenv();

        let partial_match = get_bool_from_env_or("PARTIAL_MATCH", true).unwrap();
        let key_prefix = get_env_or("KEY_PREFIX", "".to_string()).unwrap();
        let max_results = get_env_or("MAX_RESULTS", 500).unwrap();

        Self {
            partial_match,
            max_results,
            key_prefix,
        }
    }
}

#[derive(Debug, Clone)]
pub struct AppConfig {
    pub ip: String,
    pub port: u16,

    pub database_url: String,
    pub pool_size: u32,

    pub redis_url: String,

    pub static_dir: String,
    pub static_url: String,

    pub max_body_size: u64,

    pub posts_per_page: u32,

    pub upload_config: UploadConfig,
    pub search_config: SearchConfig,
}

impl AppConfig {
    pub fn from_env() -> Self {
        load_dotenv();

        let ip = get_env_or("IP", "127.0.0.1".to_string()).unwrap();
        let port = get_env_or("PORT", 8000).unwrap();

        let database_url = env::var("DATABASE_URL").expect("DATABASE_URL must be set");
        let pool_size = get_env_or("DATABASE_POOL_SIZE", 10).unwrap();

        let redis_url = get_env_or("REDIS_URL", "redis://127.0.0.1".to_string()).unwrap();

        let static_dir = get_env_or("STATIC_DIR", "./static".to_string()).unwrap();
        let static_url = get_env_or("STATIC_URL", "/static".to_string()).unwrap();

        let max_body_size = get_size_from_env_or("MAX_BODY_SIZE", 8 * 1024 * 1024).unwrap();
        let posts_per_page = get_env_or("POSTS_PER_PAGE", 30).unwrap();

        AppConfig {
            ip,
            port,

            database_url,
            pool_size,

            redis_url,

            static_dir,
            static_url,

            max_body_size,

            posts_per_page,

            upload_config: UploadConfig::from_env(),
            search_config: SearchConfig::from_env(),
        }
    }
}

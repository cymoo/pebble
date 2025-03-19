use crate::config::db::DB;
use crate::config::rd::RD;
use crate::config::AppConfig;
use crate::errors::{any_error, ApiError};
use crate::route::{post_api, post_page};
use crate::service::search_service::FullTextSearch;
use axum::extract::DefaultBodyLimit;
use axum::handler::HandlerWithoutStateExt;
use axum::http::Uri;
use axum::response::{IntoResponse, Response};
use axum::Router;
use jieba_rs::Jieba;
use std::fs;
use std::sync::Arc;
use tower::ServiceBuilder;
use tower_http::catch_panic::CatchPanicLayer;
use tower_http::cors::{Any, CorsLayer};
use tower_http::services::ServeDir;
use tower_http::trace::TraceLayer;
use tracing::error;

pub mod config;
pub mod route;
pub mod errors;
pub mod model;
pub mod util;
pub mod middleware;
pub mod service;

pub async fn create_app(state: AppState) -> Router {
    let config = &state.config;

    let static_route = Router::new()
        .nest_service(
            &config.static_url,
            ServeDir::new(config.static_dir.clone())
                .not_found_service(handle_404.into_service()),
        );

    fs::create_dir_all(config.upload_config.upload_dir.clone())
        .expect("Failed to create 'uploads' directory");

    let uploads_route = Router::new()
        .nest_service(
            &config.upload_config.upload_url,
            ServeDir::new(config.upload_config.upload_dir.clone())
                .not_found_service(handle_404.into_service()),
        );

    let max_body_size = config.max_body_size;

    // The order of the layers is important.
    // https://docs.rs/axum/latest/axum/middleware/index.html#ordering
    Router::new()
        .nest("/api", post_api::create_routes(state.rd.pool.clone()))
        .nest("/shared", post_page::create_routes())
        .merge(static_route)
        .merge(uploads_route)
        .fallback(handle_404)
        .method_not_allowed_fallback(handle_405)
        .layer(ServiceBuilder::new()
            .layer(CatchPanicLayer::custom(handle_panic))
            .layer(TraceLayer::new_for_http())
            // NOTE: Middleware added with Router::layer will run after routing
            // https://stackoverflow.com/questions/75355826/route-paths-with-or-without-of-trailing-slashes-in-rust-axum
            // https://www.matsimitsu.com/blog/2023-07-30-trailing-slashes-for-axum-routes
            // .layer(NormalizePathLayer::trim_trailing_slash())
            .layer(DefaultBodyLimit::max(max_body_size as usize))
            .layer(CorsLayer::new().allow_origin(Any).allow_methods(Any).allow_headers(Any))
        )
        .with_state(state)
}

pub async fn handle_404(_uri: Uri) -> ApiError {
    any_error(404, "Not Found", None)
}

async fn handle_405() -> ApiError {
    any_error(405, "Method Not Allowed", None)
}

fn handle_panic(panic: Box<dyn std::any::Any + Send>) -> Response {
    let panic_message = if let Some(s) = panic.downcast_ref::<&str>() {
        *s
    } else if let Some(s) = panic.downcast_ref::<String>() {
        s.as_str()
    } else {
        "Unknown panic"
    };

    error!("App panicked: {}", panic_message);
    any_error(500, "Internal Server Error", None).into_response()
}

#[derive(Clone)]
pub struct AppState {
    pub config: Arc<AppConfig>,
    pub db: Arc<DB>,
    pub rd: Arc<RD>,
    pub searcher: Arc<FullTextSearch>,
}

impl AppState {
    pub async fn new() -> Self {
        let config = AppConfig::from_env();

        let db = Arc::new(DB::new(
            &config.database_url,
            config.pool_size,
        ).await.expect("Cannot connect to database"));

        let rd = Arc::new(RD::new(&config.redis_url)
            .await.expect("Cannot connect to redis server"));

        let searcher = Arc::new(FullTextSearch::new(
            rd.clone(),
            Arc::new(Jieba::new()),
            config.search_config.partial_match,
            config.search_config.max_results,
            config.search_config.key_prefix.clone(),
        ));

        AppState {
            config: Arc::new(config),
            db,
            rd: rd.clone(),
            searcher,
        }
    }

    pub fn clone(&self) -> Self {
        Self {
            config: self.config.clone(),
            db: self.db.clone(),
            rd: self.rd.clone(),
            searcher: self.searcher.clone(),
        }
    }
}


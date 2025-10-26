use crate::config::rd::RedisPool;
use crate::errors::{not_found, ApiError, ApiResult};
use crate::middleware::check_access::check_access;
use crate::middleware::limit_request::limit_request;
use crate::model::post::{
    CreatePostRequest, CreateResponse, DateRange, DeletePostRequest, FileInfo, FilterPostRequest,
    Id, LoginRequest, Name, Post, PostPagination, PostStats, SearchRequest, UpdatePostRequest,
};
use crate::model::tag::{RenameTagRequest, StickyTagRequest, Tag, TagWithPostCount};
use crate::service::auth_service::AuthService;
use crate::service::upload_service::FileUploadService;
use crate::util::common::{to_datetime, Pipe};
use crate::util::extractor::{Json, Query, ValidatedJson, ValidatedQuery};
use crate::AppState;
use axum::extract::{Multipart, State};
use axum::http::StatusCode;
use axum::response::Html;
use axum::routing::{get, post};
use axum::{middleware, Router};
use std::collections::HashMap;
use regex::Regex;
use std::borrow::Cow;
use tracing::error;

pub fn create_routes(rd_pool: RedisPool) -> Router<AppState> {
    Router::new()
        .route("/get-tags", get(get_tags))
        .route("/rename-tag", post(rename_tag))
        .route("/stick-tag", post(stick_tag))
        .route("/delete-tag", post(delete_tag))
        .route("/search", get(search_posts))
        .route("/get-posts", get(get_posts))
        .route("/get-post", get(get_post))
        .route("/create-post", post(create_post))
        .route("/update-post", post(update_post))
        .route("/delete-post", post(delete_post))
        .route("/restore-post", post(restore_post))
        .route("/clear-posts", post(clear_posts))
        .route("/get-overall-counts", get(get_stats))
        .route("/get-daily-post-counts", get(get_daily_post_counts))
        .route("/upload", get(file_form).post(upload_file))
        .route(
            "/_dangerously_rebuild_all_indexes",
            get(rebuild_all_indexes),
        )
        .route("/auth", get(|| async {}))
        .route(
            "/login",
            post(login).layer(middleware::from_fn(move |req, next| {
                limit_request(rd_pool.clone(), 60, 5, req, next)
            })),
        )
        .layer(middleware::from_fn(|req, next| {
            check_access(&["/login"], req, next)
        }))
}

async fn login(Json(payload): Json<LoginRequest>) -> ApiResult<StatusCode> {
    if AuthService::is_valid_token(&payload.password) {
        Ok(StatusCode::NO_CONTENT)
    } else {
        Err(ApiError::Unauthorized("wrong password".to_string()))
    }
}

async fn get_tags(State(state): State<AppState>) -> ApiResult<Json<Vec<TagWithPostCount>>> {
    let tags = Tag::get_all_with_post_count(&state.db).await?;
    Ok(Json(tags))
}

async fn rename_tag(
    State(state): State<AppState>,
    Json(tag): Json<RenameTagRequest>,
) -> ApiResult<StatusCode> {
    Tag::rename_or_merge(&state.db, &tag.name, &tag.new_name).await?;
    Ok(StatusCode::NO_CONTENT)
}

async fn delete_tag(
    State(state): State<AppState>,
    Json(name): Json<Name>,
) -> ApiResult<StatusCode> {
    Tag::delete_associated_posts(&state.db, &name.name).await?;
    Ok(StatusCode::NO_CONTENT)
}

async fn stick_tag(
    State(state): State<AppState>,
    Json(tag): Json<StickyTagRequest>,
) -> ApiResult<StatusCode> {
    Tag::insert_or_update(&state.db, &tag.name, tag.sticky).await?;
    Ok(StatusCode::NO_CONTENT)
}

async fn get_posts(
    State(state): State<AppState>,
    Query(query): Query<FilterPostRequest>,
) -> ApiResult<Json<PostPagination>> {
    let posts = Post::filter_posts(&state.db, &query, 30).await?;
    let size = posts.len() as i64;
    let cursor = if size == 0 {
        -1
    } else {
        posts.last().unwrap().row.created_at
    };
    Json(PostPagination {
        posts,
        cursor,
        size,
    })
    .pipe(Ok)
}

async fn get_post(State(state): State<AppState>, Query(query): Query<Id>) -> ApiResult<Json<Post>> {
    let post = Post::find_with_parent(&state.db, query.id).await?;
    Ok(Json(post))
}

async fn search_posts(
    State(state): State<AppState>,
    ValidatedQuery(query): ValidatedQuery<SearchRequest>,
) -> ApiResult<Json<PostPagination>> {
    let (tokens, results) = state
        .fts
        .search(
            query.query.as_str(),
            query.partial.unwrap_or(false),
            query.limit.unwrap_or(0),
        )
        .await?;
    if results.is_empty() {
        return Ok(Json(PostPagination {
            posts: vec![],
            cursor: -1,
            size: 0,
        }));
    }
    let id_to_score: HashMap<i64, f64> = results.into_iter().map(|r| (r.0, r.1)).collect();
    let ids: Vec<i64> = id_to_score.keys().cloned().collect();

    let mut posts = Post::find_by_ids(&state.db, &ids).await?;

    for post in posts.iter_mut() {
        let score = id_to_score[&post.row.id];
        post.row.content = mark_tokens_in_html(&post.row.content, &tokens);
        post.score = Some(score);
    }

    posts.sort_by(|a, b| {
        b.score
            .partial_cmp(&a.score)
            .unwrap_or(std::cmp::Ordering::Equal)
    });
    let size = posts.len() as i64;

    Json(PostPagination {
        posts,
        cursor: -1,
        size,
    })
    .pipe(Ok)
}

async fn create_post(
    State(state): State<AppState>,
    ValidatedJson(post): ValidatedJson<CreatePostRequest>,
) -> ApiResult<Json<CreateResponse>> {
    let content = post.content.clone();
    let res = Post::create(&state.db, &post).await?;

    tokio::spawn(async move {
        let rv = state.fts.index(res.id, &content).await;
        if rv.is_err() {
            error!("Cannot index post: {:?}", rv);
        }
    });
    Ok(Json(res))
}

async fn update_post(
    State(state): State<AppState>,
    Json(post): Json<UpdatePostRequest>,
) -> ApiResult<StatusCode> {
    let record = Post::find_by_id(&state.db, post.id).await?;

    record
        .filter(|p| p.deleted_at.is_none())
        .ok_or_else(|| not_found("Post not found"))?;

    Post::update(&state.db, &post).await?;

    if post.content.is_present() {
        tokio::spawn(async move {
            let rv = state.fts.reindex(post.id, post.content.get()).await;
            if rv.is_err() {
                error!("Cannot rebuild index: {:?}", rv);
            }
        });
    }

    Ok(StatusCode::NO_CONTENT)
}

async fn delete_post(
    State(state): State<AppState>,
    Json(payload): Json<DeletePostRequest>,
) -> ApiResult<StatusCode> {
    if payload.hard {
        Post::clear(&state.db, payload.id).await?;

        tokio::spawn(async move {
            let rv = state.fts.deindex(payload.id).await;
            if rv.is_err() {
                error!("Cannot delete index: {:?}", rv);
            }
        });
    } else {
        Post::delete(&state.db, payload.id).await?;
    }
    Ok(StatusCode::NO_CONTENT)
}

async fn clear_posts(State(state): State<AppState>) -> ApiResult<StatusCode> {
    let ids = Post::clear_all(&state.db).await?;

    tokio::spawn(async move {
        for id in ids {
            let rv = state.fts.deindex(id).await;
            if rv.is_err() {
                error!("Cannot delete index: {:?}", rv);
                break;
            }
        }
    });

    Ok(StatusCode::NO_CONTENT)
}

async fn restore_post(
    State(state): State<AppState>,
    Json(payload): Json<Id>,
) -> ApiResult<StatusCode> {
    Post::restore(&state.db, payload.id).await?;
    Ok(StatusCode::NO_CONTENT)
}

async fn get_stats(State(state): State<AppState>) -> ApiResult<Json<PostStats>> {
    Json(PostStats {
        post_count: Post::get_count(&state.db).await?,
        tag_count: Tag::get_count(&state.db).await?,
        day_count: Post::get_active_days(&state.db).await?,
    })
    .pipe(Ok)
}

async fn get_daily_post_counts(
    State(state): State<AppState>,
    ValidatedQuery(query): ValidatedQuery<DateRange>,
) -> ApiResult<Json<Vec<i64>>> {
    Json(
        Post::get_daily_counts(
            &state.db,
            to_datetime(&query.start_date, query.offset, false)?,
            to_datetime(&query.end_date, query.offset, true)?,
        )
        .await?,
    )
    .pipe(Ok)
}

// For quick test
async fn file_form() -> Html<&'static str> {
    Html(
        r#"
        <!doctype html>
        <html>
            <head><title>Upload file</title></head>
            <body>
                <form action="upload" method="post" enctype="multipart/form-data">
                    <input type="file" name="file" multiple>
                    <button type="submit">Upload</button>
                </form>
            </body>
        </html>
        "#,
    )
}

async fn upload_file(
    State(state): State<AppState>,
    mut multipart: Multipart,
) -> ApiResult<Json<FileInfo>> {
    if let Some(field) = multipart.next_field().await? {
        let upload_service = FileUploadService::new(state.config.upload.clone());
        let rv = upload_service.stream_to_file(field).await?;
        Ok(Json(rv))
    } else {
        Err(ApiError::BadRequest("Invalid Multipart".into()))?
    }
}

async fn rebuild_all_indexes(State(state): State<AppState>) -> ApiResult<&'static str> {
    let posts = sqlx::query!("SELECT id, content FROM posts")
        .fetch_all(&state.db.pool)
        .await?;

    tokio::spawn(async move {
        let rv = state.fts.clear_all_indexes().await;
        if rv.is_err() {
            error!("Cannot clear indexes: {:?}", rv);
            return;
        }

        for post in posts {
            let rv = state.fts.index(post.id, &post.content).await;
            if rv.is_err() {
                error!("Cannot rebuild index: {:?}", rv);
                break;
            }
        }
    });

    Ok("Indexing...")
}

// Helper functions

/// Check if a character is Chinese
fn is_chinese_character(c: char) -> bool {
    ('\u{4e00}'..='\u{9fff}').contains(&c)
}

/// Convert a token into a regex pattern, handling word boundaries for Latin text
fn token_to_pattern(token: &str) -> String {
    let escaped = regex::escape(token);
    if token.chars().any(is_chinese_character) {
        escaped
    } else if token.chars().all(|c| c.is_ascii_alphanumeric() || c.is_ascii_whitespace()) {
        // Only add word boundaries for pure alphanumeric tokens
        format!(r"\b{}\b", escaped)
    } else {
        // For tokens with special characters, use the escaped version without word boundaries
        escaped
    }
}

/// Mark all occurrences of tokens in HTML text with <mark> tags,
/// avoiding replacements in HTML tags and their attributes.
///
/// # Arguments
/// * `html` - The HTML text to process
/// * `tokens` - List of tokens to be marked
///
/// # Returns
/// The processed HTML text with tokens marked
pub fn mark_tokens_in_html(html: &str, tokens: &[String]) -> String {
    if tokens.is_empty() {
        return html.to_string();
    }

    // Sort tokens by length in descending order
    let mut sorted_tokens = tokens.to_vec();
    sorted_tokens.sort_by(|a, b| b.len().cmp(&a.len()));

    // Create the regex pattern
    let patterns: Vec<String> = sorted_tokens
        .iter()
        .map(|token| token_to_pattern(token))
        .collect();

    // Combine patterns with HTML tag pattern
    let pattern = format!(r"(<[^>]*>)|({})", patterns.join("|"));
    let re = Regex::new(&pattern).expect(&format!("Failed to compile regex from: {}", pattern));

    // Process the text
    let result = re.replace_all(html, |caps: &regex::Captures| {
        if caps.get(1).is_some() {
            // HTML tag matched - return unchanged
            // Convert to owned string to avoid lifetime issues
            Cow::Owned(caps[1].to_string())
        } else {
            // Token matched - wrap with mark tags
            Cow::Owned(format!("<mark>{}</mark>", &caps[0]))
        }
    });

    result.into_owned()
}


#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_empty_tokens() {
        let text = "<html><body>Hello 世界</body></html>";
        let tokens: Vec<String> = vec![];
        let result = mark_tokens_in_html(text, &tokens);
        assert_eq!(result, text);
    }

    #[test]
    fn test_plain_text() {
        let text = "Hello 世界";
        let tokens = vec!["Hello".to_string(), "世界".to_string()];
        let result = mark_tokens_in_html(text, &tokens);
        assert_eq!(result, "<mark>Hello</mark> <mark>世界</mark>");
    }

    #[test]
    fn test_html_tags() {
        let text = "<html><body>Hello 世界</body></html>";
        let tokens = vec!["body".to_string()];
        let result = mark_tokens_in_html(text, &tokens);
        assert_eq!(result, "<html><body>Hello 世界</body></html>");
    }

    #[test]
    fn test_chinese_characters() {
        let text = "这是一个测试";
        let tokens = vec!["测试".to_string()];
        let result = mark_tokens_in_html(text, &tokens);
        assert_eq!(result, "这是一个<mark>测试</mark>");
    }

    #[test]
    fn test_english_characters() {
        let text = "Hello world, this is a test.";
        let tokens = vec!["test".to_string(), "world".to_string()];
        let result = mark_tokens_in_html(text, &tokens);
        assert_eq!(result, "Hello <mark>world</mark>, this is a <mark>test</mark>.");
    }

    #[test]
    fn test_mixed_content() {
        let text = "<html><body>Hello 世界, this is a 测试.</body></html>";
        let tokens = vec!["Hello".to_string(), "世界".to_string(), "test".to_string()];
        let result = mark_tokens_in_html(text, &tokens);
        assert_eq!(
            result,
            "<html><body><mark>Hello</mark> <mark>世界</mark>, this is a 测试.</body></html>"
        );
    }

    #[test]
    fn test_repeated_tokens() {
        let text = "Hello Hello 世界 世界";
        let tokens = vec!["Hello".to_string(), "世界".to_string()];
        let result = mark_tokens_in_html(text, &tokens);
        assert_eq!(
            result,
            "<mark>Hello</mark> <mark>Hello</mark> <mark>世界</mark> <mark>世界</mark>"
        );
    }

    #[test]
    fn test_special_characters() {
        let text = "This is a test with special characters: *test*.";
        let tokens = vec!["*test*".to_string()];
        let result = mark_tokens_in_html(text, &tokens);
        assert_eq!(
            result,
            "This is a test with special characters: <mark>*test*</mark>."
        );
    }

    #[test]
    fn test_overlapping_tokens() {
        let text = "This is a test with overlapping tokens.";
        let tokens = vec!["test".to_string(), "with overlapping".to_string()];
        let result = mark_tokens_in_html(text, &tokens);
        assert_eq!(
            result,
            "This is a <mark>test</mark> <mark>with overlapping</mark> tokens."
        );
    }
}

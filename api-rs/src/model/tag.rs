use serde::{Deserialize, Serialize};
use sqlx::FromRow;

#[derive(Debug, Serialize, FromRow)]
pub struct Tag {
    pub id: i64,
    pub name: String,
    pub sticky: bool,
    pub created_at: i64,
    pub updated_at: i64,
}

#[derive(Debug, Serialize, FromRow)]
pub struct TagWithPostCount {
    pub name: String,
    pub sticky: bool,
    pub post_count: i64,
}

#[derive(Debug, Deserialize)]
pub struct RenameTagRequest {
    pub name: String,
    pub new_name: String,
}

#[derive(Debug, Deserialize)]
pub struct StickyTagRequest {
    pub name: String,
    pub sticky: bool,
}

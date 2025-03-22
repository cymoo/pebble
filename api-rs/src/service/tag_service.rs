use crate::errors::{bad_request, ApiResult};
use crate::model::post::PostRow;
use crate::model::tag::{Tag, TagWithPostCount};
use chrono::Utc;
use sqlx::{query, query_as, Sqlite, SqlitePool, Transaction};
use std::cmp::Reverse;

impl Tag {
    pub async fn get_count(pool: &SqlitePool) -> ApiResult<i64> {
        let count = query!(
            "SELECT COUNT(*) as count FROM tags"
        ).fetch_one(pool).await?.count;

        Ok(count)
    }

    pub async fn get_all_with_post_count(pool: &SqlitePool) -> ApiResult<Vec<TagWithPostCount>> {
        let tags = query_as!(
            TagWithPostCount,
            r#"
            WITH tag_posts AS (
                SELECT t.name AS tag_name, p.id AS post_id
                FROM tags t
                JOIN tag_post_assoc tpa ON t.id = tpa.tag_id
                JOIN posts p ON tpa.post_id = p.id
                WHERE p.deleted_at IS NULL
            )
            SELECT t.name AS name, 
                   t.sticky AS sticky,
                   COUNT(DISTINCT tp.post_id) AS post_count
            FROM tags t
            LEFT JOIN tag_posts tp ON tp.tag_name = t.name OR tp.tag_name LIKE (t.name || '/%')
            GROUP BY t.name
            "#
        ).fetch_all(pool).await?;

        Ok(tags)
    }

    // It will be useful in tests
    #[allow(dead_code)]
    pub async fn get_posts(pool: &SqlitePool, name: &str) -> ApiResult<Vec<PostRow>> {
        let name_like = format!("{}/%", name);
        let posts = query_as!(
            PostRow,
            r#"
            SELECT p.*
            FROM posts p
            WHERE EXISTS (
                SELECT 1
                FROM tags t
                JOIN tag_post_assoc tp ON t.id = tp.tag_id
                WHERE tp.post_id = p.id
                AND (t.name = ? OR t.name LIKE ?))
            AND p.deleted_at IS NULL
            "#,
            name,
            name_like,
        ).fetch_all(pool).await?;

        Ok(posts)
    }

    pub async fn find_or_create(tx: &mut Transaction<'_, Sqlite>, name: &str) -> ApiResult<Tag> {
        let tag = if let Some(tag) = Tag::find_by_name(tx, &name).await? {
            tag
        } else {
            Tag::create(tx, &name).await?
        };
        Ok(tag)
    }

    pub async fn stick(pool: &SqlitePool, name: &str, sticky: bool) -> ApiResult<()> {
        let now = Utc::now().timestamp_millis();

        sqlx::query!(
            r#"
            INSERT INTO tags (name, sticky, created_at, updated_at)
            VALUES (?, ?, ?, ?)
            ON CONFLICT(name) DO UPDATE SET
                sticky = excluded.sticky,
                updated_at = excluded.updated_at
            "#,
            name,
            sticky,
            now,
            now,
        ).execute(pool).await?;

        Ok(())
    }

    pub async fn delete_associated_posts(pool: &SqlitePool, name: &str) -> ApiResult<()> {
        let now = Utc::now().timestamp_millis();
        let name_like = format!("{}/%", name);

        sqlx::query!(
            r#"
            UPDATE posts
            SET deleted_at = ?1
            WHERE id IN (
                SELECT post_id
                FROM tag_post_assoc
                WHERE tag_id IN (
                    SELECT id
                    FROM tags
                    WHERE name = ?2 OR name LIKE ?3
                )
            )
            "#,
            now,
            name,
            name_like
        ).execute(pool).await?;

        Ok(())
    }

    async fn find_by_name(tx: &mut Transaction<'_, Sqlite>, name: &str) -> ApiResult<Option<Self>> {
        let tag = query_as!(
            Tag,
            "SELECT * FROM tags WHERE name = ?",
            name,
        ).fetch_optional(&mut **tx).await?;

        Ok(tag)
    }

    async fn create(tx: &mut Transaction<'_, Sqlite>, name: &str) -> ApiResult<Tag> {
        let now = Utc::now().timestamp_millis();

        let id = query!(
            r#"
            INSERT INTO tags (name, sticky, created_at, updated_at)
            VALUES (?, false, ?, ?)
            RETURNING id
            "#,
            name,
            now,
            now
        ).fetch_one(&mut **tx).await?.id;

        Ok(Tag {
            id,
            name: name.to_string(),
            sticky: false,
            created_at: now,
            updated_at: now,
        })
    }

    /// Rename a tag, and if the new tag already exists, merge the tags.
    /// Handles all descendant tags recursively with optimal performance.
    pub async fn rename_or_merge(pool: &SqlitePool, name: &str, new_name: &str) -> ApiResult<()> {
        if name == new_name {
            return Ok(());
        }

        // Check for invalid hierarchy
        if new_name.starts_with(name) && new_name.matches('/').count() > name.matches('/').count() {
            return Err(bad_request(
                &format!(r#"Cannot move "{}" to a subtag of itself "{}""#, name, new_name)
            ));
        }

        let name_like = format!("{}/%", name);

        // Get all affected tags in a single query
        let mut affected_tags = sqlx::query_as!(
            Tag,
            r#"
            SELECT * FROM tags
            WHERE name = ? OR name = ? OR name LIKE ?
            "#,
            name,
            new_name,
            name_like
        ).fetch_all(pool).await?;

        let mut tx = pool.begin().await?;

        // Split into source tag, target tag and descendants
        let source_tag = if let Some(tag) = affected_tags.iter().find(|t| t.name == name) {
            tag
        } else {
            let new_tag = Tag::create(&mut tx, name).await?;
            affected_tags.push(new_tag);
            affected_tags.last().unwrap()
        };

        let target_tag = affected_tags.iter().find(|t| t.name == new_name);

        let mut descendants: Vec<_> = affected_tags
            .iter()
            .filter(|t| t.name != name && t.name != new_name)
            .collect();
        descendants.sort_by_key(|t| Reverse(t.name.matches('/').count()));

        for descendant in descendants {
            let new_descendant_name = descendant.name.replace(name, new_name);
            let target_descendant = Tag::find_by_name(&mut tx, &new_descendant_name).await?;

            if let Some(target_descendant) = target_descendant {
                // Target exists - merge
                Tag::merge(&mut tx, descendant, &target_descendant).await?;
            } else {
                // Target doesn't exist - rename
                Tag::rename(&mut tx, descendant, &new_descendant_name).await?;
            }
        }

        if let Some(target_tag) = target_tag {
            Tag::merge(&mut tx, source_tag, target_tag).await?;
        } else {
            Tag::rename(&mut tx, source_tag, new_name).await?;
        }

        tx.commit().await?;
        Ok(())
    }

    /// Rename a tag and update all related post content.
    async fn rename(
        tx: &mut Transaction<'_, Sqlite>,
        tag: &Tag,
        new_name: &str,
    ) -> ApiResult<()> {
        let now = Utc::now().timestamp_millis();

        // Update tag name
        sqlx::query!(
            r#"
            UPDATE tags
            SET name = ?, updated_at = ?
            WHERE id = ?
            "#,
            new_name,
            now,
            tag.id
        ).execute(&mut **tx).await?;

        let source_pattern = format!(">#{}<", tag.name);
        let target_pattern = format!(">#{}<", new_name);

        // Update post content using a single query
        sqlx::query!(
            r#"
            UPDATE posts
            SET content = REPLACE(content, ?, ?)
            WHERE id IN (
                SELECT post_id
                FROM tag_post_assoc
                WHERE tag_id = ?
            )
            "#,
            source_pattern,
            target_pattern,
            tag.id
        ).execute(&mut **tx).await?;

        Ok(())
    }

    /// Merge one tag into another, updating all related posts
    async fn merge(
        tx: &mut Transaction<'_, Sqlite>,
        source_tag: &Tag,
        target_tag: &Tag,
    ) -> ApiResult<()> {
        // Update post content
        let source_pattern = format!(">#{}<", source_tag.name);
        let target_pattern = format!(">#{}<", target_tag.name);

        sqlx::query!(
            r#"
            UPDATE posts
            SET content = REPLACE(content, ?, ?)
            WHERE id IN (
                SELECT post_id
                FROM tag_post_assoc
                WHERE tag_id = ?
            )
            "#,
            source_pattern,
            target_pattern,
            source_tag.id
        ).execute(&mut **tx).await?;

        // Insert new tag associations (ignoring if they already exist)
        sqlx::query!(
            r#"
            INSERT OR IGNORE INTO tag_post_assoc (post_id, tag_id)
            SELECT post_id, ? as tag_id
            FROM tag_post_assoc
            WHERE tag_id = ?
            "#,
            target_tag.id,
            source_tag.id
        ).execute(&mut **tx).await?;

        // Delete old tag associations
        sqlx::query!(
            r#"
            DELETE FROM tag_post_assoc
            WHERE tag_id = ?
            "#,
            source_tag.id
        ).execute(&mut **tx).await?;

        Ok(())
    }
}

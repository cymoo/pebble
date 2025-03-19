use crate::AppState;
use chrono::{Duration, Local, Utc};
use std::error::Error;
use tokio_cron_scheduler::{Job, JobScheduler};
use tracing::info;

pub async fn start_jobs(state: AppState) -> Result<(), Box<dyn Error + Send + Sync>> {
    let clear_deleted_posts = Job::new_async_tz("0 0 3 * * *", Local, move |_uuid, _l| {
        let db = state.db.pool.clone();

        Box::pin(async move {
            info!("[Daily] Checking the posts to be deleted...");

            let sixty_days_ago = (Utc::now() - Duration::days(60)).timestamp_millis();
            let rv = sqlx::query!(
                "DELETE FROM posts WHERE deleted_at < $1",
                sixty_days_ago,
            ).execute(&db).await.ok();

            if let Some(rv) = rv {
                if rv.rows_affected() > 0 {
                    info!("[Daily] Successfully deleted {} posts", rv.rows_affected());
                }
            }
        })
    })?;

    let sched = JobScheduler::new().await?;
    sched.add(clear_deleted_posts).await?;
    sched.start().await?;

    Ok(())
}

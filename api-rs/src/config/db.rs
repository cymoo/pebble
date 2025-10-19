use ::anyhow::Result;
use sqlx::sqlite::SqlitePoolOptions;
use sqlx::sqlite::{SqliteConnectOptions, SqliteJournalMode, SqlitePool};
use std::ops::Deref;
use std::str::FromStr;

pub struct DB {
    pub pool: SqlitePool,
}

impl DB {
    // Create a single connection pool for SQLx that is shared across the entire application.
    // This prevents the need to open a new connection for every API call, which would be wasteful.
    pub async fn new(url: &str, pool_size: u32) -> Result<Self> {
        let opts = SqliteConnectOptions::from_str(url)?
            .journal_mode(SqliteJournalMode::Wal)
            .foreign_keys(true);

        let pool = SqlitePoolOptions::new()
            .max_connections(pool_size)
            .connect_with(opts)
            .await?;
        Ok(DB { pool })
    }

    pub async fn migrate(&self) -> Result<()> {
        sqlx::migrate!("./migrations").run(&self.pool).await?;
        Ok(())
    }
}

impl Deref for DB {
    type Target = SqlitePool;

    fn deref(&self) -> &Self::Target {
        &self.pool
    }
}

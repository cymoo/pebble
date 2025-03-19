use bb8::Pool;
use bb8_redis::RedisConnectionManager;
use std::ops::Deref;

pub type RedisPool = Pool<RedisConnectionManager>;

pub struct RD {
    pub pool: RedisPool,
}

impl RD {
    pub async fn new(url: &str) -> anyhow::Result<Self> {
        let redis_manager = RedisConnectionManager::new(url)?;
        let redis_pool = Pool::builder().build(redis_manager).await?;

        Ok(RD { pool: redis_pool })
    }
}

impl Deref for RD {
    type Target = RedisPool;

    fn deref(&self) -> &Self::Target {
        &self.pool
    }
}

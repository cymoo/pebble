use crate::config::rd::RD;
use bb8::PooledConnection;
use bb8_redis::RedisConnectionManager;
use redis::{AsyncCommands, FromRedisValue, Pipeline, ToRedisArgs};
use serde::{Deserialize, Serialize};
use std::collections::HashSet;
use std::hash::Hash;

impl RD {
    pub async fn get_connection(
        &self,
    ) -> anyhow::Result<PooledConnection<'_, RedisConnectionManager>> {
        let conn = self.pool.get().await?;
        Ok(conn)
    }

    pub async fn exists<K: ToRedisArgs + Send + Sync>(&self, key: K) -> anyhow::Result<bool> {
        let mut conn = self.get_connection().await?;
        let exists: bool = conn.exists(key).await?;
        Ok(exists)
    }

    pub async fn scard<K: ToRedisArgs + Send + Sync>(&self, key: K) -> anyhow::Result<usize> {
        let mut conn = self.get_connection().await?;
        let count: usize = conn.scard(key).await?;
        Ok(count)
    }

    pub async fn get<T, K: ToRedisArgs + Send + Sync>(&self, key: K) -> anyhow::Result<Option<T>>
    where
        T: FromRedisValue,
    {
        let mut conn = self.get_connection().await?;
        let value: Option<T> = conn.get(key).await?;
        Ok(value)
    }

    pub async fn get_object<T, K: ToRedisArgs + Send + Sync>(
        &self,
        key: K,
    ) -> anyhow::Result<Option<T>>
    where
        T: for<'de> Deserialize<'de>,
    {
        let mut conn = self.get_connection().await?;
        let json: Option<String> = conn.get(key).await?;
        match json {
            Some(data) => Ok(Some(serde_json::from_str(&data)?)),
            None => Ok(None),
        }
    }

    pub async fn mget<T: FromRedisValue, K: ToRedisArgs + Send + Sync>(
        &self,
        key: K,
    ) -> anyhow::Result<Vec<Option<T>>> {
        let mut conn = self.get_connection().await?;
        let values: Vec<Option<T>> = conn.mget(key).await?;
        Ok(values)
    }

    pub async fn mget_object<T, K: ToRedisArgs + Send + Sync>(
        &self,
        key: K,
    ) -> anyhow::Result<Vec<Option<T>>>
    where
        T: for<'de> Deserialize<'de>,
    {
        let mut conn = self.get_connection().await?;
        let json: Vec<Option<String>> = conn.mget(key).await?;
        let mut result: Vec<Option<T>> = Vec::new();

        for opt_string in json {
            match opt_string {
                Some(data) => result.push(Some(serde_json::from_str(&data)?)),
                None => result.push(None),
            }
        }

        Ok(result)
    }

    //noinspection DuplicatedCode
    pub async fn set<T, K: ToRedisArgs + Send + Sync>(
        &self,
        key: K,
        value: T,
        expire_seconds: Option<u64>,
    ) -> anyhow::Result<()>
    where
        T: ToRedisArgs + Send + Sync,
    {
        let mut conn = self.get_connection().await?;
        if let Some(expire) = expire_seconds {
            conn.set_ex::<_, _, ()>(key, value, expire).await?;
        } else {
            conn.set::<_, _, ()>(key, value).await?;
        }
        Ok(())
    }

    //noinspection DuplicatedCode
    pub async fn set_object<T, K: ToRedisArgs + Send + Sync>(
        &self,
        key: K,
        value: &T,
        expire_seconds: Option<u64>,
    ) -> anyhow::Result<()>
    where
        T: Serialize,
    {
        let mut conn = self.get_connection().await?;
        let json = serde_json::to_string(value)?;
        if let Some(expire) = expire_seconds {
            conn.set_ex::<_, _, ()>(key, json, expire).await?;
        } else {
            conn.set::<_, _, ()>(key, json).await?;
        }
        Ok(())
    }

    pub async fn incr<K: ToRedisArgs + Send + Sync>(&self, key: K) -> anyhow::Result<i64> {
        let mut conn = self.get_connection().await?;
        let value: i64 = conn.incr(key, 1).await?;
        Ok(value)
    }

    pub async fn decr<K: ToRedisArgs + Send + Sync>(&self, key: K) -> anyhow::Result<i64> {
        let mut conn = self.get_connection().await?;
        let value: i64 = conn.decr(key, 1).await?;
        Ok(value)
    }

    pub async fn smembers<T, K: ToRedisArgs + Send + Sync>(
        &self,
        key: K,
    ) -> anyhow::Result<HashSet<T>>
    where
        T: FromRedisValue + Hash + Eq,
    {
        let mut conn = self.get_connection().await?;
        let members: HashSet<T> = conn.smembers(key).await?;
        Ok(members)
    }

    pub async fn keys<K: ToRedisArgs + Send + Sync>(
        &self,
        pattern: K,
    ) -> anyhow::Result<Vec<String>> {
        let mut conn = self.get_connection().await?;
        let keys: Vec<String> = conn.keys(pattern).await?;
        Ok(keys)
    }

    pub async fn del<K: ToRedisArgs + Send + Sync>(&self, key: K) -> anyhow::Result<()> {
        let mut conn = self.get_connection().await?;
        conn.del::<_, ()>(key).await?;
        Ok(())
    }

    pub async fn pipeline<T, F>(&self, callback: F) -> anyhow::Result<T>
    where
        F: FnOnce(&mut Pipeline),
        T: FromRedisValue,
    {
        let mut conn = self.get_connection().await?;
        let mut pipe = redis::pipe();
        let pipe = pipe.atomic();
        callback(pipe);
        let rv = pipe.query_async(&mut *conn).await;

        rv.map_err(anyhow::Error::from)
    }
}

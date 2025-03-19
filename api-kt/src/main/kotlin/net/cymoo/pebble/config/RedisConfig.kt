package net.cymoo.pebble.config

import io.lettuce.core.RedisClient
import io.lettuce.core.RedisURI
import io.lettuce.core.api.StatefulRedisConnection
import io.lettuce.core.support.ConnectionPoolSupport
import jakarta.annotation.PreDestroy
import org.apache.commons.pool2.impl.GenericObjectPool
import org.apache.commons.pool2.impl.GenericObjectPoolConfig
import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.context.annotation.Bean
import java.time.Duration

typealias RedisPool = GenericObjectPool<StatefulRedisConnection<String, String>>

@ConfigurationProperties("spring.data.redis")
data class RedisConfig(
    val host: String,
    val port: Int,
    val database: Int,
    val timeout: Long,
    val maxTotal: Int,
    val maxIdle: Int,
) {
    private lateinit var redisClient: RedisClient
    private lateinit var redisPool: RedisPool

    @Bean
    fun redisClient(): RedisClient {
        val uri =
            RedisURI.Builder.redis(host, port)
                .withDatabase(database)
                .withTimeout(Duration.ofSeconds(timeout))
                .build()
        return RedisClient.create(uri).also { redisClient = it }
    }

    @Bean
    fun redisPool(redisClient: RedisClient): RedisPool {
        val poolConfig = GenericObjectPoolConfig<StatefulRedisConnection<String, String>>().apply {
            maxTotal = maxTotal
            maxIdle = maxIdle
            // Validate connection before returning it to the connection pool to avoid putting a stale connection back into the pool.
            // This has a minor performance overhead.
            testOnReturn = true
            // Periodically check if idle connections are valid, automatically remove invalid idle connections to prevent long-idle connections from becoming stale.
            testWhileIdle = true
            // https://github.com/redis/jedis/issues/2781#issuecomment-1032632503
            jmxEnabled = false
        }
        return ConnectionPoolSupport.createGenericObjectPool({ redisClient.connect() }, poolConfig)
            .also { redisPool = it }
    }

    @PreDestroy
    fun destroy() {
        redisPool.close()
        redisClient.shutdown()
    }
}

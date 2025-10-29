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
    val url: String,
    val timeout: Long,
    val maxTotal: Int,
    val maxIdle: Int,
) {
    private lateinit var client: RedisClient
    private lateinit var pool: RedisPool

    @Bean
    fun redisClient(): RedisClient {
        val uri = RedisURI.create(url)
        uri.timeout = Duration.ofSeconds(timeout)
        return RedisClient.create(uri).also { client = it }
    }

    @Bean
    fun redisPool(redisClient: RedisClient): RedisPool {
        val poolConfig = GenericObjectPoolConfig<StatefulRedisConnection<String, String>>().apply {
            maxTotal = this@RedisConfig.maxTotal
            maxIdle = this@RedisConfig.maxIdle
            testOnReturn = true
            testWhileIdle = true
            jmxEnabled = false
        }
        return ConnectionPoolSupport.createGenericObjectPool(
            { redisClient.connect() },
            poolConfig
        ).also { pool = it }
    }

    @PreDestroy
    fun destroy() {
        if (::pool.isInitialized) {
            pool.close()
        }
        if (::client.isInitialized) {
            client.shutdown()
        }
    }
}

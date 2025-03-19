package net.cymoo.pebble.service

import com.fasterxml.jackson.core.type.TypeReference
import com.fasterxml.jackson.databind.ObjectMapper
import io.lettuce.core.LettuceFutures
import io.lettuce.core.RedisFuture
import io.lettuce.core.ScanArgs
import io.lettuce.core.ScanCursor
import io.lettuce.core.api.async.RedisAsyncCommands
import io.lettuce.core.api.sync.RedisCommands
import net.cymoo.pebble.config.RedisPool
import org.springframework.stereotype.Service
import java.time.Duration

@Service
@Suppress("unused")
class RedisService(
    val redisPool: RedisPool,
    val objectMapper: ObjectMapper
) {
    fun set(key: String, value: String): String = executeSync { it.set(key, value) }

    fun get(key: String): String? = executeSync { it.get(key) }

    fun del(vararg keys: String): Long = executeSync { it.del(*keys) }

    fun incr(key: String): Long = executeSync { it.incr(key) }

    fun decr(key: String): Long = executeSync { it.decr(key) }

    fun hset(key: String, field: String, value: String): Boolean =
        executeSync { it.hset(key, field, value) }

    fun hmset(key: String, map: Map<String, String>): String =
        executeSync { it.hmset(key, map) }

    fun hget(key: String, field: String): String? =
        executeSync { it.hget(key, field) }

    fun hgetAll(key: String): Map<String, String> =
        executeSync { it.hgetall(key) }

    fun hdel(key: String, vararg fields: String): Long =
        executeSync { it.hdel(key, *fields) }

    fun sadd(key: String, vararg members: String): Long =
        executeSync { it.sadd(key, *members) }

    fun smembers(key: String): Set<String> =
        executeSync { it.smembers(key) }

    fun srem(key: String, vararg members: String): Long =
        executeSync { it.srem(key, *members) }

    fun scard(key: String): Long =
        executeSync { it.scard(key) }

    fun exists(key: String): Boolean = executeSync { it.exists(key) != 0L }

    fun keys(pattern: String): Set<String> = executeSync { it.keys(pattern).toSet() }

    fun mget(keys: List<String>): List<String?> {
        return executeSync { it.mget(*keys.toTypedArray()) }
            .map { if (it.hasValue()) it.value else null }
    }

    final inline fun <reified T : Any> mgetObject(keys: List<String>): List<T?> {
        val typeReference = object : TypeReference<T>() {}
        return executeSync { it.mget(*keys.toTypedArray()) }
            .map {
                if (it.hasValue()) objectMapper.readValue(it.value, typeReference)
                else null
            }
    }

    final inline fun <reified T> getObject(key: String): T? {
        val typeReference = object : TypeReference<T>() {}
        return get(key)?.let {
            objectMapper.readValue(it, typeReference)
        }
    }

    fun <T : Any> setObject(key: String, value: T) {
        val json = objectMapper.writeValueAsString(value)
        set(key, json)
    }

    fun deleteByPrefix(prefix: String, batchSize: Long = 100): Long {
        return executeSync {
            var deletedCount = 0L
            var cursor = ScanCursor.INITIAL
            val scanArgs = ScanArgs.Builder.matches("$prefix*").limit(batchSize)

            do {
                val scanResult = it.scan(cursor, scanArgs)
                val keys = scanResult.keys

                if (keys.isNotEmpty()) {
                    val deleted = it.del(*keys.toTypedArray())
                    deletedCount += deleted
                }

                cursor = scanResult
            } while (!cursor.isFinished)

            deletedCount
        }
    }

    fun multi(callback: (RedisCommands<String, String>).() -> Any) {
        redisPool.borrowObject().use { conn ->
            val commands = conn.sync()
            try {
                commands.multi()
                commands.callback()
                commands.exec()
            } catch (e: Exception) {
                commands.discard()
                throw RuntimeException("Failed to execute Redis command", e)
            }
        }
    }

    // https://github.com/redis/lettuce/wiki/Pipelining-and-command-flushing
    fun <R> pipeline(
        timeout: Duration = Duration.ofSeconds(5),
        callback: RedisAsyncCommands<String, String>.() -> List<RedisFuture<R>>
    ): List<R> {
        return redisPool.borrowObject().use { conn ->
            try {
                // disable auto-flushing
                conn.setAutoFlushCommands(false)
                val futures = conn.async().callback()
                // write all commands to the transport layer
                conn.flushCommands()
                // Wait until all futures complete
                val success = LettuceFutures.awaitAll(timeout, *futures.toTypedArray())
                if (!success) {
                    throw RuntimeException("Pipeline execution timed out after ${timeout.seconds} seconds")
                }
                futures.map { it.get() }
            } catch (e: Exception) {
                throw RuntimeException("Failed to execute Redis commands", e)
            } finally {
                conn.setAutoFlushCommands(true)
            }
        }
    }

    final inline fun <R> executeSync(crossinline callback: (RedisCommands<String, String>) -> R): R {
        return try {
            redisPool.borrowObject().use { conn ->
                conn.setAutoFlushCommands(true)
                callback(conn.sync())
            }
        } catch (e: Exception) {
            throw RuntimeException("Failed to execute Redis command", e)
        }
    }
}

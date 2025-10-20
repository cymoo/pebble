package net.cymoo.pebble.config

import jakarta.annotation.PostConstruct
import net.cymoo.pebble.logger
import org.jooq.DSLContext
import org.springframework.context.annotation.Profile
import org.springframework.stereotype.Component

@Component
@Profile("!test")
class DatabaseConfigValidator(
    private val dslContext: DSLContext
) {

    @PostConstruct
    fun validateDatabaseConfig() {
        try {
            val journalMode = dslContext.fetchValue("PRAGMA journal_mode") as String
            val foreignKeys = dslContext.fetchValue("PRAGMA foreign_keys") as Int

            logger.info("Database configuration validation:")
            logger.info("Journal Mode: $journalMode")
            logger.info("Foreign Keys: $foreignKeys")

            // Validate WAL mode is enabled
            if (!journalMode.equals("wal", ignoreCase = true)) {
                throw IllegalStateException("WAL mode is not enabled. Current mode: $journalMode")
            }

            // Validate foreign key constraints are enabled
            if (foreignKeys != 1) {
                throw IllegalStateException("Foreign key constraints are not enabled")
            }
        } catch (e: Exception) {
            logger.error("Database configuration validation failed", e)
            throw e
        }
    }
}

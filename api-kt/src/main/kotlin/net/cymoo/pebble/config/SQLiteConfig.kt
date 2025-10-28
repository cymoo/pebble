package net.cymoo.pebble.config

import jakarta.annotation.PostConstruct
import net.cymoo.pebble.logger
import org.jooq.DSLContext
import org.springframework.context.annotation.Configuration
import org.springframework.context.annotation.Profile
import org.springframework.transaction.annotation.Transactional

@Configuration
@Profile("!test")
class SQLiteConfig(
    private val dsl: DSLContext
) {

    @PostConstruct
    @Transactional
    fun configureSQLite() {
        dsl.execute("PRAGMA journal_mode = WAL")
        dsl.execute("PRAGMA foreign_keys = ON")

        val journalMode = dsl.fetchValue("PRAGMA journal_mode") as String
        val foreignKeys = dsl.fetchValue("PRAGMA foreign_keys") as Int

        logger.info("SQLite journal Mode: $journalMode")
        logger.info("SQLite Foreign Keys: $foreignKeys")
    }
}

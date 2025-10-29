package site.daydream.mote.service

import org.flywaydb.core.Flyway
import org.springframework.stereotype.Service

@Service
class MigrationService(private val flyway: Flyway) {

    fun getMigrationInfo(): List<MigrationInfo> {
        return flyway.info().all().map { migration ->
            MigrationInfo(
                version = migration.version.version,
                description = migration.description,
                type = migration.type.name(),
                installedOn = migration.installedOn?.toString(),
                state = migration.state.name
            )
        }
    }

    fun repair() {
        flyway.repair()
    }

    fun migrate() {
        flyway.migrate()
    }
}

data class MigrationInfo(
    val version: String,
    val description: String,
    val type: String,
    val installedOn: String?,
    val state: String
)

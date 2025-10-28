package net.cymoo.pebble.util

import io.github.cdimascio.dotenv.Dotenv
import net.cymoo.pebble.logger
import java.io.File

object Env {
    fun load(requiredVars: List<String> = emptyList()) {
        // 1. Read system variable or default profile
        val activeProfile = System.getenv("SPRING_PROFILES_ACTIVE")
            ?: System.getProperty("spring.profiles.active")
            ?: "prod"

        // Set to system property if not already set
        if (System.getProperty("spring.profiles.active") == null) {
            System.setProperty("spring.profiles.active", activeProfile)
        }

        logger.info("Active profile: $activeProfile")

        // 2. Define loading order (priority: low to high)
        val filesToLoad = listOf(
            ".env",
            ".env.$activeProfile",
            ".env.local"
        )

        // 3. Load in order (later files override earlier ones)
        val loadedVars = mutableMapOf<String, String>()
        filesToLoad.forEach { fileName ->
            val envFile = resolveEnvFile(fileName)
            if (envFile.exists()) {
                loadDotenvFile(envFile, loadedVars)
            } else {
                logger.debug("Skipping missing env file: ${envFile.absolutePath}")
            }
        }

        // Apply loaded variables (respect system env priority)
        loadedVars.forEach { (key, value) ->
            if (System.getenv(key) == null) {
                System.setProperty(key, value)
            }
        }

        // 4. Check required variables
        requiredVars.forEach { checkRequired(it) }
    }

    fun get(key: String): String? {
        return System.getenv(key) ?: System.getProperty(key)
    }

    private fun resolveEnvFile(fileName: String): File {
        // Try project root first, then current directory
        val projectRoot = System.getProperty("user.dir")
        return File(projectRoot, fileName)
    }

    private fun loadDotenvFile(file: File, accumulator: MutableMap<String, String>) {
        try {
            val dotenv = Dotenv.configure()
                .directory(file.parent)
                .filename(file.name)
                .ignoreIfMalformed()
                .ignoreIfMissing()
                .load()

            dotenv.entries().forEach { entry ->
                accumulator[entry.key] = entry.value  // Later files override
            }
            logger.info("Loaded environment file: ${file.name} (${dotenv.entries().size} entries)")
        } catch (e: Exception) {
            logger.warn("Failed to load ${file.name}: ${e.message}")
        }
    }

    private fun checkRequired(key: String) {
        val value = System.getenv(key) ?: System.getProperty(key)
        if (value.isNullOrBlank()) {
            throw IllegalStateException("Missing required environment variable: $key")
        }
    }
}

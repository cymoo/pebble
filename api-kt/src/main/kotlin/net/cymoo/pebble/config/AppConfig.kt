package net.cymoo.pebble.config

import jakarta.annotation.PostConstruct
import net.cymoo.pebble.logger
import org.springframework.boot.context.properties.ConfigurationProperties

@ConfigurationProperties(prefix = "app")
data class AppConfig(
    val postsPerPage: Int,
    val aboutUrl: String,
    val search: SearchConfig
) {
    data class SearchConfig(
        val keyPrefix: String,
    )

    @PostConstruct
    fun test() {
        logger.warn("config: $this")
    }
}

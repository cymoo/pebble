package site.daydream.mote.config

import jakarta.annotation.PostConstruct
import site.daydream.mote.logger
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
}

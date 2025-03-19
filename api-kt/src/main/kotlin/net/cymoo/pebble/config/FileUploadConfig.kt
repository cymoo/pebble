package net.cymoo.pebble.config

import jakarta.annotation.PostConstruct
import org.springframework.boot.context.properties.ConfigurationProperties
import java.nio.file.Files
import java.nio.file.Paths
import kotlin.io.path.exists

@ConfigurationProperties(prefix = "app.upload")
data class FileUploadConfig(
    val uploadUrl: String,
    val uploadDir: String,
    val thumbnailSize: Int,
    val imageFormats: List<String>
) {
    @PostConstruct
    fun init() {
        val path = Paths.get(uploadDir)
        if (!path.exists()) {
            Files.createDirectories(path)
        }
    }
}

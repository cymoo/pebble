package net.cymoo.pebble.config

import net.cymoo.pebble.logger
import org.springframework.boot.context.event.ApplicationReadyEvent
import org.springframework.context.annotation.Configuration
import org.springframework.context.event.EventListener
import javax.imageio.ImageIO

@Configuration
class ImageIOConfig {
    @EventListener(ApplicationReadyEvent::class)
    fun initializeWebP() {
        try {
            ImageIO.scanForPlugins()
            // Optional: Verify whether WebP support has been loaded correctly
            val readers = ImageIO.getReaderMIMETypes()
            if (readers.contains("image/webp")) {
                logger.info("WebP reader found in registered readers")
            } else {
                logger.warn("WebP reader not found in registered readers. Available formats: ${readers.joinToString()}")
            }
        } catch (e: Exception) {
            logger.warn("Failed to initialize WebP support", e)
            // Can choose to throw or continue running
            // throw RuntimeException("Failed to initialize WebP support", e)
        }
    }
}

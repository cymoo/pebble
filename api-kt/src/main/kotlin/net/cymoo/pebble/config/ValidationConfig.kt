package net.cymoo.pebble.config

import jakarta.validation.Validation
import jakarta.validation.Validator
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration

@Configuration
class ValidationConfig {
    @Bean
    fun validator(): Validator {
        return Validation.byDefaultProvider()
            .configure()
            .addValueExtractor(net.cymoo.pebble.util.maybe.MaybeMissingValueExtractor())
            .buildValidatorFactory()
            .validator
    }
}

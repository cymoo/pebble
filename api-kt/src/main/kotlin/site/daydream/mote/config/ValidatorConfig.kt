package site.daydream.mote.config

import jakarta.validation.Validation
import jakarta.validation.Validator
import site.daydream.mote.util.maybe.MaybeMissingValueExtractor
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration

@Configuration
class ValidatorConfig {
    @Bean
    fun validator(): Validator {
        return Validation.byDefaultProvider()
            .configure()
            .addValueExtractor(MaybeMissingValueExtractor())
            .buildValidatorFactory()
            .validator
    }
}

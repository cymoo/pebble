package net.cymoo.pebble.config

import com.fasterxml.jackson.databind.DeserializationFeature
import com.fasterxml.jackson.databind.MapperFeature
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.PropertyNamingStrategies
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule
import com.fasterxml.jackson.module.kotlin.KotlinFeature
import com.fasterxml.jackson.module.kotlin.KotlinModule
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration


@Configuration
class JacksonConfig {
    @Bean
    fun objectMapper(): ObjectMapper {
        return ObjectMapper().apply {
            // support kotlin features
            registerModule(
                KotlinModule.Builder()
                    .enable(KotlinFeature.NullToEmptyCollection)
                    // .enable(KotlinFeature.NullIsSameAsDefault)
                    .enable(KotlinFeature.StrictNullChecks)
                    .build()
            )
            // support Java 8 datetime
            registerModule(JavaTimeModule())
            registerModule(net.cymoo.pebble.util.maybe.MaybeMissingModule())

            propertyNamingStrategy = PropertyNamingStrategies.SNAKE_CASE

            configure(DeserializationFeature.FAIL_ON_NULL_FOR_PRIMITIVES, true)
            configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, true)
            configure(MapperFeature.ACCEPT_CASE_INSENSITIVE_ENUMS, true)
        }
    }
}

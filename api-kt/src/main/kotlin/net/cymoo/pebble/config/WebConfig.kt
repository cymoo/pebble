package net.cymoo.pebble.config

import net.cymoo.pebble.interceptor.AuthInterceptor
import net.cymoo.pebble.service.AuthService
import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.boot.convert.ApplicationConversionService
import org.springframework.context.annotation.Configuration
import org.springframework.format.FormatterRegistry
import org.springframework.web.servlet.config.annotation.CorsRegistry
import org.springframework.web.servlet.config.annotation.InterceptorRegistry
import org.springframework.web.servlet.config.annotation.ResourceHandlerRegistry
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer

@Configuration
// @EnableWebMvc
class WebConfig(
    private val cors: CorsProperties,
    private val uploadConfig: FileUploadConfig,
    private val authService: AuthService
) : WebMvcConfigurer {
    override fun addCorsMappings(registry: CorsRegistry) {
        val (origins, methods, headers, allowCredentials, maxAge) = cors

        registry.addMapping("/**")
            .allowedOrigins(*origins.toTypedArray())
            .allowedMethods(*methods.toTypedArray())
            .allowedHeaders(*headers.toTypedArray())
            .allowCredentials(allowCredentials)
            // No need to preflight (send OPTIONS request) within a certain period of time
            .maxAge(maxAge)

    }

    // https://www.baeldung.com/spring-mvc-static-resources
    override fun addResourceHandlers(registry: ResourceHandlerRegistry) {
        val (uploadUrl, uploadDir) = uploadConfig
        registry.addResourceHandler("/$uploadUrl/**")
            .addResourceLocations("file:$uploadDir/")
    }

    // https://www.baeldung.com/spring-boot-enum-mapping
    override fun addFormatters(registry: FormatterRegistry) {
        super.addFormatters(registry)
        // https://stackoverflow.com/questions/50231233/deserialize-enum-ignoring-case-in-spring-boot-controller
        // `ApplicationConversionService` comes with a set of configured converters and formatters.
        // It includes a converter that converts a string into an enum in a case-insensitive manner.

        // NOTE: This ONLY applies to query strings or form data. The case-insensitive feature for JSON needs additional configuration.
        ApplicationConversionService.configure(registry)
    }

    @Override
    override fun addInterceptors(registry: InterceptorRegistry) {
        registry.addInterceptor(AuthInterceptor(authService))
    }
}

@ConfigurationProperties(prefix = "app.cors")
data class CorsProperties(
    val allowedOrigins: List<String>,
    val allowedMethods: List<String>,
    val allowedHeaders: List<String>,
    val allowCredentials: Boolean,
    val maxAge: Long
)

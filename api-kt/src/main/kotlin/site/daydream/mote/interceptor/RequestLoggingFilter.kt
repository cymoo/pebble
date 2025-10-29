package site.daydream.mote.interceptor

import jakarta.servlet.FilterChain
import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import org.slf4j.LoggerFactory
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.stereotype.Component
import org.springframework.util.AntPathMatcher
import org.springframework.web.filter.OncePerRequestFilter
import org.springframework.web.util.ContentCachingResponseWrapper
import kotlin.math.log10
import kotlin.math.pow

@ConfigurationProperties(prefix = "logging.request")
data class RequestLoggingProperties(
    val enabled: Boolean = false,
    val excludePaths: List<String> = listOf(
        "/actuator/**",
        "/static/**",
        "/health",
        "/favicon.ico"
    )
)

@Component
@ConditionalOnProperty(
    prefix = "logging.request",
    name = ["enabled"],
    havingValue = "true",
    matchIfMissing = true
)
class RequestLoggingFilter(
    private val properties: RequestLoggingProperties
) : OncePerRequestFilter() {

    private val httpLogger = LoggerFactory.getLogger(javaClass)
    private val pathMatcher = AntPathMatcher()

    override fun shouldNotFilter(request: HttpServletRequest): Boolean {
        val path = request.servletPath
        return properties.excludePaths.any { pattern ->
            pathMatcher.match(pattern, path)
        }
    }

    override fun doFilterInternal(
        request: HttpServletRequest,
        response: HttpServletResponse,
        filterChain: FilterChain
    ) {
        val startTime = System.nanoTime()
        val responseWrapper = ContentCachingResponseWrapper(response)

        try {
            filterChain.doFilter(request, responseWrapper)
        } finally {
            logRequest(request, responseWrapper, startTime)
            responseWrapper.copyBodyToResponse()
        }
    }

    private fun logRequest(
        request: HttpServletRequest,
        response: ContentCachingResponseWrapper,
        startTime: Long
    ) {
        val duration = System.nanoTime() - startTime
        val method = request.method
        val protocol = request.protocol
        val url = buildUrl(request)
        val remoteAddr = getRemoteAddress(request)
        val remotePort = request.remotePort
        val status = response.status
        val contentLength = response.contentSize.toLong()

        val formattedSize = formatBytes(contentLength)
        val formattedDuration = formatDuration(duration)

        httpLogger.info("\"$method $url $protocol\" from $remoteAddr:$remotePort - $status $formattedSize in $formattedDuration")
    }

    private fun buildUrl(request: HttpServletRequest): String {
        val url = StringBuilder()
            .append(request.scheme)
            .append("://")
            .append(request.serverName)

        val port = request.serverPort
        if ((request.scheme == "http" && port != 80) ||
            (request.scheme == "https" && port != 443)
        ) {
            url.append(":").append(port)
        }

        url.append(request.contextPath).append(request.servletPath)

        request.queryString?.let { url.append("?").append(it) }

        return url.toString()
    }

    private fun getRemoteAddress(request: HttpServletRequest): String {
        return request.getHeader("X-Forwarded-For")
            ?.split(",")?.first()?.trim()
            ?: request.getHeader("X-Real-IP")
            ?: request.remoteAddr
    }

    private fun formatBytes(bytes: Long): String {
        if (bytes == 0L) return "0B"

        val units = arrayOf("B", "KB", "MB", "GB", "TB")
        val digitGroups = (log10(bytes.toDouble()) / log10(1024.0)).toInt()
            .coerceIn(0, units.size - 1)

        return if (digitGroups == 0) {
            "${bytes}B"
        } else {
            val value = bytes / 1024.0.pow(digitGroups.toDouble())
            "%.1f%s".format(value, units[digitGroups])
        }
    }

    private fun formatDuration(nanos: Long): String {
        return when {
            nanos < 1_000 -> "${nanos}ns"
            nanos < 1_000_000 -> "%.3fμs".format(nanos / 1_000.0)
            nanos < 1_000_000_000 -> "%.3fms".format(nanos / 1_000_000.0)
            else -> "%.3fs".format(nanos / 1_000_000_000.0)
        }
    }
}

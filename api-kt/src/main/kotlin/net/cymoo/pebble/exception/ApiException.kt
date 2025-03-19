package net.cymoo.pebble.exception

import com.fasterxml.jackson.annotation.JsonInclude

open class ApiException(
    val code: Int,
    val error: String,
    override val message: String? = null,
) : RuntimeException(message)

class NotFoundException(message: String?) :
    ApiException(404, "Not Found", message)

class BadRequestException(message: String?) :
    ApiException(400, "Bad Request", message)

class AuthenticationException(message: String?) :
    ApiException(401, "Unauthorized", message)


data class ErrorResponse(
    val code: Int,
    val error: String,
    @JsonInclude(JsonInclude.Include.NON_NULL)
    val message: String? = null,
    // val path: String,
    // @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
    // val timestamp: LocalDateTime = LocalDateTime.now()
)


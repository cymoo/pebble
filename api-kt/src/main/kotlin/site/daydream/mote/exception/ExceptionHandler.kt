package site.daydream.mote.exception

import com.fasterxml.jackson.core.JsonParseException
import com.fasterxml.jackson.databind.JsonMappingException
import com.fasterxml.jackson.databind.exc.InvalidFormatException
import com.fasterxml.jackson.databind.exc.MismatchedInputException
import com.fasterxml.jackson.databind.exc.UnrecognizedPropertyException
import jakarta.servlet.http.HttpServletRequest
import jakarta.validation.ConstraintViolationException
import site.daydream.mote.util.camelToSnake
import site.daydream.mote.util.wrapWith
import org.slf4j.LoggerFactory
import org.springframework.beans.BeanUtils
import org.springframework.beans.TypeMismatchException
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import org.springframework.http.converter.HttpMessageNotReadableException
import org.springframework.validation.BindingResult
import org.springframework.validation.FieldError
import org.springframework.web.HttpRequestMethodNotSupportedException
import org.springframework.web.bind.MethodArgumentNotValidException
import org.springframework.web.bind.MissingServletRequestParameterException
import org.springframework.web.bind.annotation.ExceptionHandler
import org.springframework.web.bind.annotation.ResponseStatus
import org.springframework.web.bind.annotation.RestControllerAdvice
import org.springframework.web.method.annotation.MethodArgumentTypeMismatchException
import org.springframework.web.multipart.MaxUploadSizeExceededException
import org.springframework.web.servlet.NoHandlerFoundException
import org.springframework.web.servlet.resource.NoResourceFoundException
import java.math.BigDecimal
import java.time.LocalDate
import java.time.LocalDateTime

@RestControllerAdvice
class ExceptionHandler {
    private val logger = LoggerFactory.getLogger(this::class.java)

    @ExceptionHandler(APIException::class)
    fun handleApiException(
        ex: APIException,
        request: HttpServletRequest
    ): ResponseEntity<ErrorResponse> {
        val errorResponse = ErrorResponse(
            code = ex.code,
            error = ex.error,
            message = ex.message,
        )
        return ResponseEntity.status(ex.code).body(errorResponse)
    }

    // 404
    @ExceptionHandler(NoHandlerFoundException::class, NoResourceFoundException::class)
    @ResponseStatus(HttpStatus.NOT_FOUND)
    fun handleNotFoundException(
        ex: Exception,
        request: HttpServletRequest
    ) = ErrorResponse(code = 404, error = "Not Found")

    // 405
    @ExceptionHandler(HttpRequestMethodNotSupportedException::class)
    @ResponseStatus(HttpStatus.METHOD_NOT_ALLOWED)
    fun handleMethodNotAllowed(
        ex: HttpRequestMethodNotSupportedException,
        request: HttpServletRequest
    ) = ErrorResponse(code = 405, error = "Method not allowed")

    // 413
    @ExceptionHandler(MaxUploadSizeExceededException::class)
    @ResponseStatus(HttpStatus.PAYLOAD_TOO_LARGE)
    fun handleMaxUploadSizeExceededException(
        ex: MaxUploadSizeExceededException,
        request: HttpServletRequest
    ) = ErrorResponse(code = 413, error = "Payload Too Large")

    // 400
    @ExceptionHandler(MethodArgumentTypeMismatchException::class)
    @ResponseStatus(HttpStatus.BAD_REQUEST)
    fun handleArgumentTypeMismatch(
        ex: MethodArgumentTypeMismatchException,
        request: HttpServletRequest
    ): ErrorResponse {
        val field = ex.name
        val targetType = ex.requiredType?.simpleName
        return ErrorResponse(
            code = 400,
            error = "Bad Request",
            message = if (targetType != null) "'$field' should be '$targetType'" else "'$field' is invalid"
        )
    }

    @ExceptionHandler(MissingServletRequestParameterException::class)
    @ResponseStatus(HttpStatus.BAD_REQUEST)
    fun handleMissingRequestParameter(
        ex: MissingServletRequestParameterException,
        request: HttpServletRequest
    ): ErrorResponse {
        return ErrorResponse(
            code = 400,
            error = "Bad Request",
            message = "'${ex.parameterName}' is required"
        )
    }

    @ExceptionHandler(MethodArgumentNotValidException::class)
    @ResponseStatus(HttpStatus.BAD_REQUEST)
    fun handleValidationException(
        ex: MethodArgumentNotValidException,
        request: HttpServletRequest
    ): ErrorResponse {
        val message = if (ex.hasGlobalErrors()) {
            ex.bindingResult.globalErrors.joinToString(", ") {
                if (it.defaultMessage == null) {
                    return@joinToString "Invalid request format"
                }

                val defaultMessage = it.defaultMessage!!

                if ("non-null is null" in defaultMessage) {
                    val fieldName = extractParameterName(defaultMessage)
                    "'$fieldName' is required"
                } else {
                    defaultMessage
                }
            }
        } else {
            ex.bindingResult.fieldErrors.joinToString(", ") {
                formatFieldError(it, ex.bindingResult)
            }
        }
        return ErrorResponse(
            code = 400,
            error = "Bad Request",
            message = message,
        )
    }

    @ExceptionHandler(ConstraintViolationException::class)
    @ResponseStatus(HttpStatus.BAD_REQUEST)
    fun handleConstraintViolation(ex: ConstraintViolationException): ErrorResponse {
        val message = ex.constraintViolations.joinToString(", ") {
            val field = it.propertyPath.joinToString(".") { node -> node.name }
            "'$field' ${it.message}"
        }
        return ErrorResponse(
            code = 400,
            error = "Bad Request",
            message = message,
        )
    }

    @ExceptionHandler(HttpMessageNotReadableException::class)
    @ResponseStatus(HttpStatus.BAD_REQUEST)
    fun handleJsonErrors(ex: HttpMessageNotReadableException): ErrorResponse {
        val errorMessage = when (val cause = ex.cause) {
            // Handle Jackson's `InvalidFormatException` (type mismatch)
            is InvalidFormatException -> handleInvalidFormat(cause)

            // Handle bad JSON format
            is JsonParseException -> "Invalid JSON"

            // Handle unknown fields
            is UnrecognizedPropertyException -> handleUnrecognizedProperty(cause)

            is MismatchedInputException -> handleMismatchedInput(cause)

            // Other JSON mapping errors
            is JsonMappingException -> handleJsonMapping(cause)

            // unknown errors
            else -> "Invalid request format"
        }

        return ErrorResponse(400, "Bad Request", errorMessage)
    }

    // 500
    @ExceptionHandler(Exception::class)
    @ResponseStatus(HttpStatus.INTERNAL_SERVER_ERROR)
    fun handleGenericException(
        ex: Exception,
        request: HttpServletRequest
    ) = ErrorResponse(
        code = 500,
        error = "Internal Server Error",
        message = "An unexpected error occurred",
    ).also {
        logger.error("Unhandled exception", ex)
    }
}

private fun formatFieldError(error: FieldError, bindingResult: BindingResult): String {
    val fieldName = "'${error.field.camelToSnake()}'"
    with(error) {
        return when (code) {
            "NotNull" -> "$fieldName cannot be null"
            "NotEmpty", "NotBlank" -> "$fieldName cannot be empty"
            "Size" -> "$fieldName size should be between ${arguments!![1]} and ${arguments!![2]}"
            "Min" -> "$fieldName should be greater than ${arguments!![1]}"
            "Max" -> "$fieldName should be less than ${arguments!![1]}"
            "Email" -> "$fieldName should be a valid email address"
            "Pattern" -> "$fieldName $defaultMessage"
            "typeMismatch" -> "$fieldName should be '${getTargetType(error, bindingResult)}'"
            else -> "$fieldName ${defaultMessage ?: "is invalid"}"
        }
    }
}

// Convert Java types to more user-friendly type names
private fun formatTypeName(type: Class<*>): String = when (type) {
    Boolean::class.java, java.lang.Boolean::class.java -> "boolean"
    Int::class.java, Integer::class.java -> "int"
    Long::class.java, java.lang.Long::class.java -> "long"
    Double::class.java, java.lang.Double::class.java -> "double"
    Float::class.java, java.lang.Float::class.java -> "float"

    LocalDate::class.java -> "Date (YYYY-MM-DD)"
    LocalDateTime::class.java -> "DateTime (YYYY-MM-DD HH:mm:ss)"
    BigDecimal::class.java -> "Decimal"
    else -> when {
        type.isEnum -> "Enum (${type.enumConstants.joinToString(", ") { it.toString().lowercase() }})"
        Collection::class.java.isAssignableFrom(type) -> "Array"
        Map::class.java.isAssignableFrom(type) -> "Object"
        type.isArray -> "Array"
        else -> type.simpleName
    }
}

private fun getTargetType(error: FieldError, bindingResult: BindingResult): String {
    val targetType = when {
        // 1. Try to obtain the target type from `TypeMismatchException`
        error.contains(TypeMismatchException::class.java) -> {
            val typeError = error.unwrap(TypeMismatchException::class.java)
            typeError.requiredType
        }
        // 2. Try to obtain the type from `BindingResult`
        else -> {
            val targetClass = bindingResult.target?.javaClass
            targetClass?.let { clazz ->
                try {
                    // Use `PropertyDescriptor` to obtain the property type
                    BeanUtils.getPropertyDescriptor(clazz, error.field)?.propertyType
                } catch (e: Exception) {
                    // 3. If fails, try to get the field type directly through reflection
                    try {
                        clazz.getDeclaredField(error.field).type
                    } catch (e: Exception) {
                        null
                    }
                }
            }
        }
    }
    return targetType?.let { formatTypeName(it) } ?: "Unknown Type"
}

private fun extractParameterName(errorMessage: String): String? {
    val parameterRegex = "parameter\\s+([a-zA-Z][a-zA-Z0-9]*)"
    val matchResult = parameterRegex.toRegex().find(errorMessage)
    return matchResult?.groupValues?.get(1)?.camelToSnake()
}

private fun buildPath(path: List<JsonMappingException.Reference>): String {
    return path.joinToString(".") { ref ->
        when {
            ref.index >= 0 -> "${ref.fieldName ?: ""}[${ref.index}]"
            else -> ref.fieldName ?: ""
        }
    }
}

private fun handleInvalidFormat(ex: InvalidFormatException): String {
    val path = buildPath(ex.path).wrapWith("'")
    val targetType = formatTypeName(ex.targetType)
    val value = ex.value?.toString() ?: "null"

    return when {
        // Special handling for enum types
        ex.targetType.isEnum -> {
            val validValues = ex.targetType.enumConstants.joinToString(", ") { it.toString().lowercase() }
            "$path should be one of: ($validValues)"
        }
        // General handling for other types
        else -> "$path should be '$targetType', but got '$value'"
    }
}

private fun handleMissingParameter(ex: MismatchedInputException): String {
    val path = buildPath(ex.path).wrapWith("'")
    return "$path is required"
}

private fun handleUnrecognizedProperty(ex: UnrecognizedPropertyException): String {
    val path = buildPath(ex.path).wrapWith("'")
    val knownProps = ex.knownPropertyIds?.joinToString(", ") ?: ""
    return if (knownProps.isNotEmpty()) {
        "Unknown field $path; Valid fields are: ($knownProps)"
    } else {
        "Unknown field $path"
    }
}

private fun handleMismatchedInput(ex: MismatchedInputException): String {
    val path = buildPath(ex.path).wrapWith("'")
    val message = ex.message ?: ""

    if (message.contains("missing", ignoreCase = true) ||
        message.contains("required", ignoreCase = true)
    ) {
        return "$path is required"
    }

    val targetType = ex.targetType
    return when {
        targetType.isArray || List::class.java.isAssignableFrom(targetType) ->
            "$path should be an array"

        targetType.kotlin.isData || Map::class.java.isAssignableFrom(targetType) ->
            "$path should be an object"

        else -> "$path is invalid"
    }
}

private fun handleJsonMapping(ex: JsonMappingException): String {
    val path = buildPath(ex.path).wrapWith("'")
    return "$path is invalid"
}

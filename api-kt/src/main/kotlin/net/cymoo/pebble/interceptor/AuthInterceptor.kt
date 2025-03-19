package net.cymoo.pebble.interceptor

import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import net.cymoo.pebble.annotation.AuthRequired
import net.cymoo.pebble.exception.AuthenticationException
import net.cymoo.pebble.service.AuthService
import org.springframework.web.method.HandlerMethod
import org.springframework.web.servlet.HandlerInterceptor

class AuthInterceptor(private val authService: AuthService) : HandlerInterceptor {
    override fun preHandle(request: HttpServletRequest, response: HttpServletResponse, handler: Any): Boolean {
        if (handler !is HandlerMethod) {
            return true
        }

        val method = handler.method
        val beanType = method.declaringClass

        val classAuth = beanType.getAnnotation(AuthRequired::class.java)
        val methodAuth = method.getAnnotation(AuthRequired::class.java)

        // Determine whether validation is required:
        // 1. If the method has an annotation, prioritize the method's configuration.
        // 2. If the method does not have an annotation but the class does, use the class's configuration.
        // 3. If neither the method nor the class has an annotation, validation is not required.
        val requiresAuth = when {
            methodAuth != null -> methodAuth.required
            classAuth != null -> classAuth.required
            else -> false
        }

        if (requiresAuth) {
            val token = getCookie(request, "token") ?: extractBearer(request)

            if (token.isNullOrEmpty()) {
                throw AuthenticationException("No token provided")
            }

            if (!authService.isValidToken(token)) {
                throw AuthenticationException("Invalid token")
            }
        }

        return true
    }
}

fun extractBearer(request: HttpServletRequest): String? {
    val authHeader = request.getHeader("Authorization")
    if (authHeader.isNullOrEmpty()) throw AuthenticationException("Missing authorization header")

    if (authHeader.startsWith("Bearer ")) {
        // Remove the "Bearer " prefix and extract the token part
        return authHeader.removePrefix("Bearer ").trim()
    }
    return null
}

fun getCookie(request: HttpServletRequest, name: String): String? {
    if (name.isEmpty()) return null

    return request.cookies?.firstOrNull { name == it.name }?.value
}

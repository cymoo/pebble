package net.cymoo.pebble.service

import org.springframework.stereotype.Service

@Service
class AuthService {
    fun isValidToken(token: String): Boolean {
        return token == System.getenv("PEBBLE_PASSWORD")
    }
}

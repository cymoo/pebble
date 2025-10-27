package net.cymoo.pebble.service

import net.cymoo.pebble.util.Env
import org.springframework.stereotype.Service

@Service
class AuthService {
    fun isValidToken(token: String): Boolean {
        return token == Env.get("PEBBLE_PASSWORD")
    }
}

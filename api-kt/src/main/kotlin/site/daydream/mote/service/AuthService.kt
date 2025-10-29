package site.daydream.mote.service

import site.daydream.mote.util.Env
import org.springframework.stereotype.Service

@Service
class AuthService {
    fun isValidToken(token: String): Boolean {
        return token == Env.get("MOTE_PASSWORD")
    }
}

package net.cymoo.pebble

import net.cymoo.pebble.util.Env
import org.junit.jupiter.api.Test
import org.springframework.boot.test.context.SpringBootTest

@SpringBootTest
class SpringApplicationTests {

    companion object {
        init {
            Env.load()
        }
    }

    @Test
    fun contextLoads() {
    }

}

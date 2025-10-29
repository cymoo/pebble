package site.daydream.mote

import org.junit.jupiter.api.Test
import org.springframework.boot.test.context.SpringBootTest
import site.daydream.mote.util.Env

@SpringBootTest
class SpringApplicationTests {

    companion object {
        init {
            System.setProperty("spring.profiles.active", "test")
            Env.load()
        }
    }

    @Test
    fun contextLoads() {
    }
}

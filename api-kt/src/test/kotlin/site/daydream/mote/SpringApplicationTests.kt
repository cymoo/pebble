package site.daydream.mote

import site.daydream.mote.util.Env
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

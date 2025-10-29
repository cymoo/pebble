package site.daydream.mote.controller

import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.http.MediaType
import org.springframework.test.web.servlet.MockMvc
import org.springframework.test.web.servlet.get
import org.springframework.test.web.servlet.post

@SpringBootTest
@AutoConfigureMockMvc
@Disabled
class PostApiControllerTest(@Autowired val mockMvc: MockMvc) {

    @Test
    fun `should create a post and assert content field`() {
        val post = """{"content":"Hello, world!"}"""

        mockMvc.post("/api/create-post") {
            contentType = MediaType.APPLICATION_JSON
            content = post
        }.andExpect {
            status { isOk() }
            content { json(post) }
        }

        mockMvc.get("/api/posts") {
        }.andExpect {
            status { isOk() }
            jsonPath("$[0].id") { value(1) }
        }
    }
}

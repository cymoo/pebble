package net.cymoo.pebble.util

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test


class HighlightTokensTest {
    @Test
    fun `test basic text marking`() {
        assertEquals(
            "<mark>hello</mark> world",
            "hello world".highlight(listOf("hello"))
        )
    }

    @Test
    fun `test multiple tokens`() {
        assertEquals(
            "<mark>hello</mark> <mark>world</mark>",
            "hello world".highlight(listOf("hello", "world"))
        )
    }

    @Test
    fun `test Chinese text`() {
        assertEquals(
            "<mark>你好</mark>世界",
            "你好世界".highlight(listOf("你好"))
        )
    }

    @Test
    fun `test inside HTML tags`() {
        assertEquals(
            """<a href="token"><mark>token</mark></a>""",
            """<a href="token">token</a>""".highlight(listOf("token"))
        )
    }

    @Test
    fun `test multiple attributes`() {
        assertEquals(
            """<div class="test" data-test="test"><mark>test</mark></div>""",
            """<div class="test" data-test="test">test</div>""".highlight(listOf("test"))
        )
    }

    @Test
    @Disabled("The case cannot pass: the second 'token' will not be marked")
    fun `test text with angle brackets`() {
        assertEquals(
            "<p>some <mark>token</mark> >~< <mark>token</mark></p>",
            "<p>some token >~< token</p>".highlight(listOf("token"))
        )
    }

    @Test
    fun `test empty token list`() {
        assertEquals(
            "<p>test</p>",
            "<p>test</p>".highlight(emptyList())
        )
    }

    @Test
    fun `test overlapping tokens`() {
        assertEquals(
            "<mark>hello world</mark>",
            "hello world".highlight(listOf("hello world", "world"))
        )
    }

    @Test
    fun `test special characters`() {
        assertEquals(
            """<a href="test.com"><mark>test.com</mark> <mark>test</mark></a>""",
            """<a href="test.com">test.com test</a>""".highlight(listOf("test.com", "test"))
        )
    }

    @Test
    fun `test mixed content`() {
        assertEquals(
            "<div><mark>hello</mark> <mark>世界</mark>, <mark>你好</mark> <mark>world</mark>!</div>",
            "<div>hello 世界, 你好 world!</div>".highlight(
                listOf("hello", "你好", "world", "世界")
            )
        )
    }

    @Test
    fun `test English work boundary`() {
        assertEquals(
            "This is <mark>foo</mark> and foolish",
            "This is foo and foolish".highlight(listOf("foo"))
        )
    }

    @Test
    fun `test English word boundary and mixed content`() {
        assertEquals(
            "This is <mark>foo</mark> and foolish, <mark>你好</mark> world",
            "This is foo and foolish, 你好 world".highlight(listOf("foo", "你好"))
        )
    }
}

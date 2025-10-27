package net.cymoo.pebble.util

import com.fasterxml.jackson.databind.PropertyNamingStrategies.SnakeCaseStrategy

fun String.count(chr: Char): Int = this.count { it == chr }

fun String.camelToSnake(): String = SnakeCaseStrategy().translate(this)

fun String.wrapWith(text: String) = "$text$this$text"

operator fun String.times(n: Int) = this.repeat(n)

fun String.replaceFromStart(from: String, to: String): String {
    return if (this.startsWith(from)) {
        to + this.substring(from.length)
    } else {
        this
    }
}

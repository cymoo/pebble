package net.cymoo.pebble.util

import com.fasterxml.jackson.databind.PropertyNamingStrategies.SnakeCaseStrategy
import java.time.LocalDate
import java.time.ZoneOffset
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter
import java.util.*
import kotlin.math.abs

/** Check if a character is Chinese */
private fun Char.isChineseCharacter(): Boolean {
    return this in '\u4e00'..'\u9fff'
}

fun String.count(chr: Char): Int = this.count { it == chr }

fun String.camelToSnake(): String = SnakeCaseStrategy().translate(this)

fun String.wrapWith(text: String) = "$text$this$text"

operator fun String.times(n: Int) = this.repeat(n)

/**
 * Convert a date string to a ZonedDateTime object with timezone information
 *
 * @param offset  Timezone offset in minutes
 * @param endOfDay Whether to use the end time of the day (defaults to false, which means the start time of the day)
 * @return A ZonedDateTime object
 * @throws IllegalArgumentException if the timezone offset is out of range
 */
fun String.toDateTime(offset: Int, endOfDay: Boolean = false): ZonedDateTime {
    require(abs(offset) <= 1440) {
        "Timezone offset must be between -1440 and 1440 minutes: $offset"
    }

    val localDate = LocalDate.parse(this, DateTimeFormatter.ofPattern("yyyy-MM-dd"))

    val localDateTime = if (endOfDay) {
        localDate.atTime(23, 59, 59, 999_000_000)
    } else {
        localDate.atStartOfDay()
    }

    val zoneOffset = ZoneOffset.ofTotalSeconds(offset * 60)

    return localDateTime.atZone(zoneOffset)
}

/**
 * Mark all occurrences of tokens in HTML text with <mark> tags,
 * avoiding replacements in HTML tags and their attributes
 *
 * @param tokens List of tokens to be marked
 * @return HTML text with tokens marked only in text content
 */
fun String.highlight(tokens: List<String>): String {
    if (tokens.isEmpty()) return this

    // Add word boundaries for English tokens
    val patterns = tokens
        .sortedByDescending { it.length }
        .map { token ->
            if (token.any { it.isChineseCharacter() }) {
                // Chinese token
                Regex.escape(token)
            } else {
                // English token with word boundaries
                "\\b${Regex.escape(token)}\\b"
            }
        }

    // Create pattern that matches either HTML tags or tokens
    val pattern = Regex("<[^>]*>|(${patterns.joinToString("|")})")

    return pattern.replace(this) { matchResult ->
        // If group1 is null, it means we matched an HTML tag (group0)
        // Otherwise, we matched a token and should wrap it with mark tags
        matchResult.groups[1]?.let { "<mark>${it.value}</mark>" } ?: matchResult.value
    }
}

val INVALID_CHARS_REGEX = Regex("[^\\w\\-.\\u4e00-\\u9fa5]+")

/**
 * Generates a secure filename with UUID suffix of specified length
 *
 * @param filename Original filename
 * @param uuidLength Length of UUID suffix (8-32 characters)
 * @return Secured filename with UUID suffix
 * @throws IllegalArgumentException if filename is blank or uuid length is invalid
 */
fun generateSecureFilename(filename: String, uuidLength: Int = 8): String {
    require(filename.isNotBlank()) { "Filename cannot be blank" }
    require(uuidLength in 8..32) {
        "UUID length must be between 8 and 32"
    }

    val sanitizedName = filename.trim().replace(INVALID_CHARS_REGEX, "_")

    val (baseName, extension) = splitFileName(sanitizedName)
    val uuid = UUID.randomUUID().toString()
        .replace("-", "")
        .take(uuidLength)

    return buildString {
        append(baseName)
        append('.')
        append(uuid)
        if (extension.isNotEmpty()) {
            append('.')
            append(extension)
        }
    }
}

/**
 * Splits a filename into base name and extension
 * Handles special cases like hidden files and multiple extensions
 */
private fun splitFileName(fileName: String): Pair<String, String> {
    // Handle hidden file starting with `.`
    if (fileName.startsWith(".")) {
        val remaining = fileName.substring(1)
        val lastDotIndex = remaining.lastIndexOf('.')
        return if (lastDotIndex < 0) {
            ".$remaining" to ""
        } else {
            ".${remaining.substring(0, lastDotIndex)}" to remaining.substring(lastDotIndex + 1)
        }
    }

    val lastDotIndex = fileName.lastIndexOf('.')
    return when {
        lastDotIndex <= 0 -> fileName to ""
        else -> fileName.substring(0, lastDotIndex) to fileName.substring(lastDotIndex + 1)
    }
}

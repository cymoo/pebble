package net.cymoo.pebble.model

import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.Pattern
import java.time.Instant

data class Tag(
    val id: Int? = null,
    val name: String,
    val sticky: Boolean = false,
    val createdAt: Long = Instant.now().toEpochMilli(),
    val updatedAt: Long = Instant.now().toEpochMilli(),
)

data class TagWithPostCount(
    val name: String,
    val sticky: Boolean,
    val postCount: Long
)

data class TagRename(
    @field:NotBlank
    val name: String,

    @field:NotBlank
    @field:Pattern(
        regexp = "^[^/\\s#][^\\s#]*+(?<!/)$",
        message = "cannot contain spaces or '#', and cannot start/end with '/'"
    )
    @field:Pattern(
        regexp = "^(?!.*?//).*+$",
        message = "cannot not contain consecutive '/'"
    )
    val newName: String
)

data class TagStick(
    @field:NotBlank
    val name: String,

    val sticky: Boolean,
)

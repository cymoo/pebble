package net.cymoo.pebble.model

import com.fasterxml.jackson.annotation.JsonIgnore
import com.fasterxml.jackson.annotation.JsonInclude
import com.fasterxml.jackson.annotation.JsonRawValue
import com.fasterxml.jackson.core.JsonGenerator
import com.fasterxml.jackson.databind.JsonSerializer
import com.fasterxml.jackson.databind.SerializerProvider
import com.fasterxml.jackson.databind.annotation.JsonSerialize
import jakarta.validation.constraints.NotBlank
import org.springframework.web.bind.annotation.BindParam
import java.time.Instant

data class Post(
    val id: Int? = null,
    val content: String,

    @JsonSerialize(nullsUsing = DefaultEmptyListSerializer::class)
    @get:JsonRawValue
    val files: String? = null,

    val color: String? = null,
    val shared: Boolean = false,

    @JsonInclude(JsonInclude.Include.NON_NULL)
    val deletedAt: Long? = null,

    val createdAt: Long = Instant.now().toEpochMilli(),
    val updatedAt: Long = Instant.now().toEpochMilli(),

    @JsonIgnore
    val parentId: Int? = null,

    @JsonInclude(JsonInclude.Include.NON_NULL)
    val parent: Post? = null,

    val childrenCount: Int = 0,

    @JsonInclude(JsonInclude.Include.NON_NULL)
    val children: List<Post>? = null,

    // Search relevance score
    @JsonInclude(JsonInclude.Include.NON_NULL)
    val score: Double? = null
)

enum class CategoryColor { RED, BLUE, GREEN }

enum class SortingField { CREATED_AT, UPDATED_AT, DELETED_AT }

data class Id(
    val id: Int
)

data class Name(
    @field:NotBlank
    val name: String
)

data class LoginRequest(
    @field:NotBlank
    val password: String,
)

data class FileInfo(
    val url: String,
    val thumbUrl: String?,
    val size: Long?,
    val width: Int?,
    val height: Int?
)

data class PostQuery(
    @field:NotBlank
    val query: String,
)

data class PostFilterOptions(
    val cursor: Long?,
    val deleted: Boolean = false,

    // Mapping query parameters to @ModelAttribute does not respect Jackson's SNAKE_CASE naming strategy.
    @BindParam("parent_id")
    val parentId: Int?,

    val color: CategoryColor?,

    val tag: String?,
    val shared: Boolean?,

    @BindParam("has_files")
    val hasFiles: Boolean?,

    @BindParam("order_by")
    val orderBy: SortingField = SortingField.CREATED_AT,

    val ascending: Boolean = false,

    @BindParam("start_date")
    val startDate: Long? = null,
    @BindParam("end_date")
    val endDate: Long? = null,
)

data class DateRange(
    @BindParam("start_date")
    @field:ValidDate
    val startDate: String,

    @BindParam("end_date")
    @field:ValidDate
    val endDate: String,

    val offset: Int = 480, // timezone offset in minutes
)

data class PostCreate(
    @field:NotBlank
    val content: String,

    val files: List<FileInfo>?,
    val color: CategoryColor?,
    val shared: Boolean?,
    val parentId: Int?,
)

data class PostUpdate(
    val id: Int,

    val content: String?,
    val shared: Boolean?,

    val files: net.cymoo.pebble.util.maybe.MaybeMissing<List<FileInfo>?>,
    val color: net.cymoo.pebble.util.maybe.MaybeMissing<CategoryColor?>,
    val parentId: net.cymoo.pebble.util.maybe.MaybeMissing<Int?>,
)

data class PostDelete(
    val id: Int,
    val hard: Boolean = false,
)

data class CreateResponse(
    val id: Int,
    val createdAt: Long,
    val updatedAt: Long,
)

data class PostPagination(
    val posts: List<Post>,
    val cursor: Long,
    val size: Int,
)

data class PostStats(
    val postCount: Int,
    val tagCount: Int,
    val dayCount: Int,
)

class DefaultEmptyListSerializer : JsonSerializer<List<*>?>() {
    override fun serialize(value: List<*>?, gen: JsonGenerator, provider: SerializerProvider) {
        gen.writeStartArray()
        gen.writeEndArray() // Return an empty array [] when null
    }
}


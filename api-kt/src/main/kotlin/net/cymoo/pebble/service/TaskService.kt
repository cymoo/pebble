package net.cymoo.pebble.service

import net.cymoo.pebble.generated.Tables.POSTS
import net.cymoo.pebble.logger
import org.jooq.DSLContext
import org.springframework.scheduling.annotation.Async
import org.springframework.scheduling.annotation.Scheduled
import org.springframework.stereotype.Service
import java.time.Instant
import java.time.temporal.ChronoUnit

@Service
class TaskService(private val searchService: SearchService, private val dsl: DSLContext) {

    @Scheduled(cron = "0 0 3 * * ?")
    fun clearPosts() {
        val thirtyDaysAgo = Instant.now().minus(30, ChronoUnit.DAYS).toEpochMilli()
        logger.info("Clearing posts deleted before: ${Instant.ofEpochMilli(thirtyDaysAgo)}")
        val deletedCount = dsl.deleteFrom(POSTS)
            .where(POSTS.DELETED_AT.lessThan(thirtyDaysAgo))
            .execute()
        if (deletedCount > 0) {
            logger.info("Successfully deleted $deletedCount posts.")
        }
    }

    @Async
    fun buildIndex(id: Int, content: String) {
        searchService.index(id, content)
    }

    @Async
    fun rebuildIndex(id: Int, content: String) {
        searchService.reindex(id, content)
    }

    @Async
    fun deleteIndex(id: Int) {
        searchService.deindex(id)
    }

    @Async
    fun rebuildAllIndexes() {
        searchService.clearAllIndexes()
        dsl.select(POSTS.ID, POSTS.CONTENT).from(POSTS).fetch().forEach {
            searchService.index(it.value1(), it.value2())
        }
    }
}

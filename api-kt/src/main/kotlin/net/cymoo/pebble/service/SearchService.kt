package net.cymoo.pebble.service

import com.fasterxml.jackson.databind.ObjectMapper
import com.hankcs.hanlp.tokenizer.IndexTokenizer
import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Service
import kotlin.math.log10
import kotlin.math.max
import kotlin.math.sqrt

data class SearchResult(
    val tokens: List<String>,
    val matches: List<DocumentMatch>
)

data class DocumentMatch(
    val id: Int,
    val score: Double
)

class TextAnalyzer {
    fun analyze(text: String): List<String> = text
        .replace(HTML_TAG, " ")
        .replace(PUNCTUATION, " ")
        .let { IndexTokenizer.segment(it) }
        .filter { it.word.trim().isNotEmpty() }
        .map { it.word.lowercase() }
        .filterNot { it in STOP_WORDS }

    companion object {
        private val HTML_TAG = Regex("<[^>]*>")

        val PUNCTUATION = "\\p{P}".toRegex()
        val STOP_WORDS = setOf(
            "a", "an", "and", "are", "as", "at", "be", "by", "can",
            "for", "from", "have", "if", "in", "is", "it",
            "may", "not", "of", "on", "or", "tbd", "that", "the",
            "this", "to", "us", "we", "when", "will", "with",
            "yet", "you", "your", "的", "了", "和", "着", "与"
        )
    }
}

@Service
class SearchService(
    private val redisService: RedisService,
    private val objectMapper: ObjectMapper,

    @Value("\${app.search.key-prefix:}")
    private val keyPrefix: String,
) {
    private val textAnalyzer = TextAnalyzer()

    fun isIndexed(id: Int): Boolean =
        redisService.exists(docTokensKey(id))

    fun getDocCount(): Int =
        redisService.get(docCountKey())?.toIntOrNull() ?: 0

    fun index(id: Int, text: String) {
        if (isIndexed(id)) {
            return reindex(id, text)
        }

        val tokens = textAnalyzer.analyze(text)
        if (tokens.isEmpty()) {
            return
        }

        val tokenFrequency = tokens.groupingBy { it }.eachCount()

        redisService.multi {
            incr(docCountKey())
            set(docTokensKey(id), objectMapper.writeValueAsString(tokenFrequency))
            for (token in tokenFrequency.keys) {
                sadd(tokenDocsKey(token), id.toString())
            }
        }
    }

    fun reindex(id: Int, text: String) {
        if (!isIndexed(id)) {
            return index(id, text)
        }

        val newTokens = textAnalyzer.analyze(text)
        if (newTokens.isEmpty()) {
            return deindex(id)
        }

        val newTokenFrequency = newTokens.groupingBy { it }.eachCount()

        val oldTokenFrequency = redisService.getObject<Map<String, Int>>(docTokensKey(id))
        require(!oldTokenFrequency.isNullOrEmpty())

        redisService.multi {
            set(docTokensKey(id), objectMapper.writeValueAsString(newTokenFrequency))
            val oldTokenSet = oldTokenFrequency.keys
            val newTokenSet = newTokenFrequency.keys

            for (token in oldTokenSet - newTokenSet) {
                srem(tokenDocsKey(token), id.toString())
            }

            for (token in newTokenSet - oldTokenSet) {
                sadd(tokenDocsKey(token), id.toString())
            }
        }
    }

    fun deindex(id: Int) {
        val tokenFrequency = redisService.getObject<Map<String, Int>>(docTokensKey(id))
        require(!tokenFrequency.isNullOrEmpty())

        redisService.multi {
            del(docTokensKey(id))
            decr(docCountKey())

            for (token in tokenFrequency.keys) {
                srem(tokenDocsKey(token), id.toString())
            }
        }
    }

    fun clearAllIndexes() {
        redisService.deleteByPrefix("${keyPrefix}doc:")
        redisService.deleteByPrefix("${keyPrefix}token:")
    }

    fun search(query: String, partial: Boolean = true, limit: Int = 0): SearchResult {
        val tokens = textAnalyzer.analyze(query)
        if (tokens.isEmpty()) {
            return SearchResult(tokens, emptyList())
        }

        val docIds = findMatchingDocuments(tokens, partial)
        if (docIds.isEmpty()) {
            return SearchResult(tokens, emptyList())
        }

        var rankedResults = rankDocuments(tokens, docIds)
            .sortedByDescending { it.score }

        if (limit > 0) {
            rankedResults = rankedResults.take(limit)
        }

        return SearchResult(tokens, rankedResults)
    }

    private fun findMatchingDocuments(tokens: List<String>, partial: Boolean): Set<Int> {
        val results = redisService.pipeline {
            tokens.map { token -> smembers(tokenDocsKey(token)) }
        }

        return results.map { result -> result.map { it.toInt() }.toSet() }
            .reduce { acc, ids ->
                if (partial) acc.union(ids) else acc.intersect(ids)
            }
    }

    private fun rankDocuments(tokens: List<String>, docIds: Set<Int>): List<DocumentMatch> {
        val totalDocs = getDocCount().toDouble()

        // Fetch term frequencies (TFs) for the given document IDs
        val tokenFrequencies = redisService.mgetObject<Map<String, Int>>(docIds.map { docTokensKey(it) })

        // Fetch document frequencies (DFs) for the query tokens
        val docFrequencies = redisService.pipeline { tokens.map { scard(tokenDocsKey(it)) } }.map { it.toDouble() }

        return docIds.zip(tokenFrequencies).map { (id, tokenFreq) ->
            var matchingTerms = 0
            var score = tokens.zip(docFrequencies).sumOf { (token, df) ->
                val tf = tokenFreq?.get(token)?.toDouble() ?: 0.0
                if (tf > 0) matchingTerms += 1

                // Use an improved TF calculation: 1 + log(tf) to reduce the weight of high-frequency terms
                val normalizedTf = if (tf > 0) 1 + log10(tf) else 0.0
                val idf = if (df > 0) log10(max(1.0, totalDocs / df)) else 0.0
                normalizedTf * idf
            }

            // Apply a length normalization factor to avoid advantages for long documents
            val totalTerms = tokenFreq?.values?.sum() ?: 0
            if (totalTerms > 0) {
                score /= sqrt(totalTerms.toFloat())
            }

            // Calculate query term coverage
            val coverageRatio = matchingTerms.toDouble() / tokens.size.toDouble()

            score *= if (coverageRatio > 0.999) {
                //  Weight documents with full query term matches
                2.0
            } else {
                coverageRatio
            }

            DocumentMatch(id, score)
        }
    }

    private fun docCountKey() = "${keyPrefix}doc:count"
    private fun docTokensKey(id: Int) = "${keyPrefix}doc:$id:tokens"
    private fun tokenDocsKey(token: String) = "${keyPrefix}token:$token:docs"
}

package site.daydream.mote.service

import org.junit.jupiter.api.*
import org.junit.jupiter.api.Assertions.*
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.test.context.ActiveProfiles

@SpringBootTest
@ActiveProfiles("test")
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class SearchServiceTest(
) {
    @Autowired
    private lateinit var searchService: SearchService

    @AfterEach
    fun setup() {
        searchService.clearAllIndexes()
    }

    @Nested
    inner class BasicIndexingTests {
        @Test
        fun `should successfully index and retrieve English document`() {
            // given
            val id = 1
            val content = "The quick brown fox jumps over the lazy dog"

            // when
            searchService.index(id, content)

            // then
            assertTrue(searchService.isIndexed(id))
            assertEquals(1, searchService.getDocCount())

            // when searching
            val searchResult = searchService.search("quick fox")

            // then
            with(searchResult) {
                assertTrue(tokens.containsAll(listOf("quick", "fox")))
                assertEquals(1, matches.size)
                assertEquals(id, matches[0].id)
            }
        }

        @Test
        fun `should successfully index and retrieve Chinese document`() {
            // given
            val id = 1
            val content = "苹果是一个伟大的公司"

            // when
            searchService.index(id, content)

            // then
            assertTrue(searchService.isIndexed(id))
            assertEquals(1, searchService.getDocCount())

            // when searching
            val searchResult = searchService.search("苹果 伟大")

            // then
            with(searchResult) {
                assertTrue(tokens.containsAll(listOf("苹果", "伟大")))
                assertEquals(1, matches.size)
                assertEquals(id, matches[0].id)
            }
        }

        @Test
        fun `should handle mixed language content`() {
            // given
            val id = 1
            val content = "Learning 中文 is very 有趣 and 有意义"

            // when
            searchService.index(id, content)

            // then
            assertTrue(searchService.isIndexed(id))

            // when searching Chinese
            val chineseResult = searchService.search("中文 有趣")

            // then
            with(chineseResult) {
                assertTrue(tokens.containsAll(listOf("中文", "有趣")))
                assertEquals(1, matches.size)
            }

            // when searching English
            val englishResult = searchService.search("learning")

            // then
            with(englishResult) {
                assertTrue(tokens.contains("learning"))
                assertEquals(1, matches.size)
            }
        }
    }

    @Nested
    inner class DocumentUpdateTests {
        @Test
        fun `should update existing document content`() {
            // given
            val id = 1
            val originalContent = "原始内容测试文档"
            searchService.index(id, originalContent)

            // when
            val newContent = "更新后的内容与测试"
            searchService.reindex(id, newContent)

            // then
            val oldSearch = searchService.search("原始")
            assertEquals(0, oldSearch.matches.size)

            val newSearch = searchService.search("更新")
            assertEquals(1, newSearch.matches.size)
        }

        @Test
        fun `should handle document deletion`() {
            // given
            val id = 1
            val content = "待删除的测试文档"
            searchService.index(id, content)

            // when
            searchService.deindex(id)

            // then
            assertFalse(searchService.isIndexed(id))
            assertEquals(0, searchService.getDocCount())

            val searchResult = searchService.search("测试")
            assertEquals(0, searchResult.matches.size)
        }
    }

    @Nested
    inner class MultipleDocumentTests {
        @Test
        fun `should rank documents by relevance`() {
            // given
            val docs = listOf(
                Pair(1, "Python 编程语言教程，人人学会编程"),
                Pair(2, "Python 是最流行的编程语言之一"),
                Pair(3, "Java 编程入门教程"),
                Pair(4, "Python 和 Java 的比较"),
                Pair(5, "Rust 有很高的安全性和性能")
            )

            // when
            docs.forEach { (id, content) ->
                searchService.index(id, content)
            }

            // then
            val pythonResults = searchService.search("Python 编程")
            with(pythonResults) {
                assertEquals(4, matches.size)
                assertTrue(matches[0].score > matches[1].score)
            }

            val javaResults = searchService.search("Java")
            assertEquals(2, javaResults.matches.size)
        }

        @Test
        fun `should handle search with multiple terms and languages`() {
            // given
            searchService.index(1, "机器学习 Machine Learning 入门")
            searchService.index(2, "深度学习 Deep Learning 实战")
            searchService.index(3, "自然语言处理 NLP learning guide")

            // when
            val result1 = searchService.search("机器学习 learning")
            val result2 = searchService.search("深度学习 deep")
            val result3 = searchService.search("nlp 教程")

            // then
            assertEquals(3, result1.matches.size)
            assertEquals(2, result2.matches.size)
            assertEquals(1, result3.matches.size)
        }
    }

    @Nested
    inner class EdgeCaseTests {
        @Test
        fun `should handle empty content`() {
            // given
            val id = 1
            val content = ""

            // when
            searchService.index(id, content)

            // then
            assertFalse(searchService.isIndexed(id))
            assertEquals(0, searchService.getDocCount())
        }

        @Test
        fun `should handle content with only stop words`() {
            // given
            val id = 1
            val content = "的 了 着 和 与 the and is"

            // when
            searchService.index(id, content)

            // then
            assertFalse(searchService.isIndexed(id))
            assertEquals(0, searchService.getDocCount())
        }

        @Test
        fun `should handle special characters and HTML content`() {
            // given
            val id = 1
            val content = """
                <div>测试文档！@#￥%……&*（）</div>
                <p>Special,.:;?/\|+=</p>
                <span>Test Content</span>
            """.trimIndent()

            // when
            searchService.index(id, content)

            // then
            assertTrue(searchService.isIndexed(id))

            val result = searchService.search("测试 test")
            assertEquals(1, result.matches.size)
        }
    }

    @Nested
    inner class PerformanceTests {
        @Test
        fun `should handle large number of documents`() {
            // given
            val documentCount = 1000
            val documents = (1..documentCount).map { id ->
                id to "文档 $id Document number $id 包含一些测试内容 with some test content"
            }

            // when
            documents.forEach { (id, content) ->
                searchService.index(id, content)
            }

            // then
            assertEquals(documentCount, searchService.getDocCount())

            // when searching
            val result = searchService.search("测试 test")

            assertTrue(result.matches.isNotEmpty())
        }

        @Test
        fun `should handle concurrent indexing and searching`() {
            // given
            val documentCount = 500
            val documents = (1..documentCount).map { id ->
                id to "并发测试文档 $id Concurrent test document $id"
            }

            // when
            documents.parallelStream().forEach { (id, content) ->
                searchService.index(id, content)
            }

            // then
            assertEquals(documentCount, searchService.getDocCount())

            // when searching concurrently
            val searchResults = listOf("并发", "测试", "concurrent", "test")
                .parallelStream()
                .map { term -> searchService.search(term) }
                .toList()

            // then
            assertTrue(searchResults.all { it.matches.isNotEmpty() })
        }
    }

    @Nested
    inner class SearchRelevanceTests {
        @Test
        fun `should prioritize exact matches over partial matches`() {
            // given
            searchService.index(1, "深度学习算法")
            searchService.index(2, "机器学习算法详解")
            searchService.index(3, "机器学习简单介绍")

            // when
            val result = searchService.search("机器学习 算法")

            // then
            assertEquals(3, result.matches.size)
            // 包含所有搜索词的文档应该排在最前面
            assertEquals(2, result.matches[0].id)
        }

        @Test
        fun `should consider term frequency in ranking`() {
            // given
            searchService.index(1, "Python Python Python 编程")
            searchService.index(2, "Python 编程教程")

            // when
            val result = searchService.search("python")

            // then
            assertEquals(2, result.matches.size)
            // 包含更多搜索词的文档应该排在前面
            assertEquals(1, result.matches[0].id)
        }
    }

    @Nested
    inner class DocumentModificationTests {
        @Test
        fun `should handle repeated indexing of same document`() {
            // given
            val id = 1
            val content1 = "第一版内容"
            val content2 = "第二版内容"
            val content3 = "第三版内容"

            // when
            searchService.index(id, content1)
            searchService.index(id, content2)
            searchService.index(id, content3)

            // then
            assertTrue(searchService.isIndexed(id))
            assertEquals(1, searchService.getDocCount())

            val result = searchService.search("第三版")
            assertEquals(1, result.matches.size)
        }

        @Test
        fun `should handle document content clearing`() {
            // given
            val id = 1
            searchService.index(id, "初始内容")

            // when
            searchService.reindex(id, "")

            // then
            assertFalse(searchService.isIndexed(id))
        }
    }
}

class TextAnalyzerTest {
    private lateinit var analyzer: TextAnalyzer

    @BeforeEach
    fun setup() {
        analyzer = TextAnalyzer()
    }

    @Nested
    inner class BasicAnalysisTests {
        @Test
        fun `should analyze simple English text`() {
            // given
            val text = "quick brown fox jumps"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertEquals(listOf("quick", "brown", "fox", "jumps"), tokens)
        }

        @Test
        fun `should analyze simple Chinese text`() {
            // given
            val text = "中国人民很伟大"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assert(listOf("中国", "人民", "伟大").all { it in tokens })
        }

        @Test
        fun `should analyze mixed language text`() {
            // given
            val text = "machine learning 机器学习 deep learning 深度学习"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assert(
                listOf("machine", "learning", "机器", "学习", "deep", "learning", "深度", "学习").all { it in tokens },
            )
        }
    }

    @Nested
    inner class HtmlProcessingTests {
        @Test
        fun `should remove HTML tags`() {
            // given
            val text = "<div>测试文本</div><p>test content</p>"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertEquals(listOf("测试", "文本", "test", "content"), tokens)
        }

        @Test
        fun `should handle nested HTML tags`() {
            // given
            val text = "<div><span>深度</span><p>学习</p></div>"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertEquals(listOf("深度", "学习"), tokens)
        }

        @Test
        fun `should handle HTML attributes`() {
            // given
            val text = """<div class="test" id="main">测试内容</div>"""

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertEquals(listOf("测试", "内容"), tokens)
        }
    }

    @Nested
    inner class PunctuationTests {
        @Test
        fun `should remove Chinese punctuation`() {
            // given
            val text = "中国，真是。太！伟大？了；看："

            // when
            val tokens = analyzer.analyze(text)

            // then
            val expectedTokens = listOf("中国", "真是", "太", "伟大", "看")
            assertEquals(expectedTokens, tokens)
        }

        @Test
        fun `should remove English punctuation`() {
            // given
            val text = "Hello, world! How? are: you; today."

            // when
            val tokens = analyzer.analyze(text)

            // then
            val expectedTokens = listOf("hello", "world", "how", "today")
            assertEquals(expectedTokens, tokens)
        }

        @Test
        fun `should handle mixed punctuation`() {
            // given
            val text = "Hello，世界! How。are。you?"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertEquals(listOf("hello", "世界", "how"), tokens)
        }
    }

    @Nested
    inner class StopWordsTests {
        @Test
        fun `should remove English stop words`() {
            // given
            val text = "this is a test of the system"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertEquals(listOf("test", "system"), tokens)
        }

        @Test
        fun `should remove Chinese stop words`() {
            // given
            val text = "这个 的 和 那个 与 了 着"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertEquals(listOf("这个", "那个"), tokens)
        }

        @Test
        fun `should remove mixed language stop words`() {
            // given
            val text = "this is 一个 test 的 system"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assert(listOf("is", "的").all { it !in tokens })
        }
    }

    @Nested
    inner class EdgeCasesTests {
        @Test
        fun `should handle empty text`() {
            // given
            val text = ""

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertTrue(tokens.isEmpty())
        }

        @Test
        fun `should handle text with only stop words`() {
            // given
            val text = "the and is of 的 了 和"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertTrue(tokens.isEmpty())
        }

        @Test
        fun `should handle text with only punctuation`() {
            // given
            val text = ",.!?;:，。！？；：、"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertTrue(tokens.isEmpty())
        }

        @Test
        fun `should handle whitespace text`() {
            // given
            val text = "   \n\t   "

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertTrue(tokens.isEmpty())
        }
    }

    @Nested
    inner class SpecialCasesTests {
        @Test
        fun `should handle repeated words`() {
            // given
            val text = "test test test 测试 测试 测试"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assertEquals(listOf("test", "test", "test", "测试", "测试", "测试"), tokens)
        }

        @Test
        fun `should handle numbers and special characters`() {
            // given
            val text = "测试123 test456 @#$%^&*"

            // when
            val tokens = analyzer.analyze(text)

            // then
            assert(listOf("测试", "123", "test", "456").all { it in tokens })
        }
    }
}

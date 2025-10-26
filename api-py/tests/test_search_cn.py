import os
import pytest


@pytest.fixture(scope='function')
def redis_client():
    """Create a test Redis client"""
    from redis import Redis

    redis_url = os.environ.get("REDIS_URL_TEST", "redis://localhost:6379/15")

    client = Redis.from_url(redis_url, decode_responses=True)
    yield client
    client.flushdb()


@pytest.fixture
def search_engine(redis_client):
    """Create a FullTextSearch instance test fixture"""
    from app.lib.search import FullTextSearch

    return FullTextSearch(redis_client, key_prefix='test:')


def test_basic_indexing(search_engine):
    """Test basic indexing functionality"""
    text = "这是一个测试文档，用于全文搜索"
    search_engine.index(1, text)

    assert search_engine.is_indexed(1) is True
    assert search_engine.doc_count == 1


def test_empty_text_indexing(search_engine):
    """Test empty text and special character indexing"""
    search_engine.index(1, "")
    search_engine.index(2, "!@#$%^&*()")

    assert search_engine.is_indexed(1) is False
    assert search_engine.is_indexed(2) is False
    assert search_engine.doc_count == 0


def test_html_text_indexing(search_engine):
    """Test HTML text indexing"""
    html_text = "<p>这是一个<strong>测试</strong>文档</p>"
    search_engine.index(1, html_text)

    tokens, results = search_engine.search("测试")
    assert len(results) > 0
    assert results[0][0] == 1


def test_reindex_document(search_engine):
    """Test document re-indexing"""
    search_engine.index(1, "原始文档内容")
    search_engine.reindex(1, "更新后的文档内容")

    tokens, results = search_engine.search("更新")
    assert len(results) > 0
    assert results[0][0] == 1


def test_deindex_document(search_engine):
    """Test document de-indexing"""
    search_engine.index(1, "要删除的文档")
    search_engine.deindex(1)

    assert search_engine.is_indexed(1) is False
    assert search_engine.doc_count == 0


def test_multi_document_search(search_engine):
    """Test multi-document search and relevance sorting"""
    documents = [
        "Python用于数据科学",
        "#Python Python是一种编程语言，它广泛用于数据分析中",
        "数据科学需要编程技能",
    ]

    for i, doc in enumerate(documents, 1):
        search_engine.index(i, doc)

    tokens, results = search_engine.search("Python 数据")

    assert len(results) > 1
    # Verify relevance sorting
    top_doc_id, top_score = results[0]
    assert top_doc_id in (1, 2)


@pytest.mark.parametrize(
    "query",
    [
        "Python编程",
        "数据科学",
        "机器学习",
    ],
)
def test_parametrized_search(search_engine, query):
    """Parameterized search test"""
    documents = [
        "Python是一种编程语言",
        "Python用于数据科学",
        "数据科学需要编程技能",
        "机器学习是人工智能的分支",
    ]

    for i, doc in enumerate(documents, 1):
        search_engine.index(i, doc)

    tokens, results = search_engine.search(query)
    assert len(tokens) > 0
    assert len(results) > 0


def test_max_results_limit(search_engine):
    """Test maximum results limit"""
    # Index a large number of documents
    for i in range(500):
        search_engine.index(i, f"测试文档 {i}")

    tokens, results = search_engine.search("测试", limit=300)
    assert len(results) <= 300  # Default max results


def test_clear_all_indexes(search_engine):
    """Test clearing all indexes"""
    documents = ["第一个文档", "第二个文档", "第三个文档"]

    for i, doc in enumerate(documents, 1):
        search_engine.index(i, doc)

    search_engine.clear_all_indexes()

    assert search_engine.doc_count == 0

    keys = search_engine.db.keys('test:*')
    assert len(keys) == 0


def test_complex_search_scenarios(search_engine):
    """Test complex search scenarios"""
    documents = [
        "机器学习是人工智能的重要分支",
        "深度学习是机器学习的子领域",
        "人工智能正在快速发展",
    ]

    for i, doc in enumerate(documents, 1):
        search_engine.index(i, doc)

    # Multi-keyword search
    tokens, results = search_engine.search("机器学习 人工智能")
    assert len(results) > 0

    # Verify result relevance
    top_doc_id, _ = results[0]
    assert documents[top_doc_id - 1] == "机器学习是人工智能的重要分支"

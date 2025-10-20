import pytest


@pytest.fixture
def redis_client():
    """Create a test Redis client"""
    from redis import Redis
    from app.config import TestConfig

    client = Redis(**TestConfig.REDIS)
    yield client
    client.flushdb()


@pytest.fixture
def search_engine(redis_client):
    """Create a FullTextSearch instance test fixture"""
    from app.lib.search import FullTextSearch

    return FullTextSearch(redis_client, key_prefix='test:')


def test_index_and_has_indexed(search_engine):
    """Test document indexing and index check functionality"""
    text = "Hello world, this is a test document"
    doc_id = 1

    # Initial state should not be indexed
    assert not search_engine.is_indexed(doc_id)

    # Perform indexing
    search_engine.index(doc_id, text)

    # Check is indexed
    assert search_engine.is_indexed(doc_id)


def test_search_basic(search_engine):
    """Test basic search functionality"""
    # Index multiple documents
    docs = {
        1: "The quick brown fox jumps over the lazy dog",
        2: "A quick brown rabbit runs fast",
        3: "Slow dogs sleep all day",
    }

    for doc_id, text in docs.items():
        search_engine.index(doc_id, text)

    # Search test
    tokens, results = search_engine.search("quick brown")
    assert tokens == ["quick", "brown"]
    assert len(results) > 0

    # Verify result order and relevance
    top_doc_id = results[0][0]
    assert top_doc_id in [1, 2]  # Should match document 1 or 2


def test_reindex(search_engine):
    """Test re-indexing functionality"""
    doc_id = 1
    original_text = "Hello world python programming"
    updated_text = "Hello python advanced programming"

    # First indexing
    search_engine.index(doc_id, original_text)

    # Re-index
    search_engine.reindex(doc_id, updated_text)

    tokens, results = search_engine.search("world")
    assert len(results) == 0

    # Search verification
    tokens, results = search_engine.search("advanced")
    assert len(results) > 0
    assert results[0][0] == doc_id


def test_deindex(search_engine):
    """Test document de-indexing functionality"""
    doc_id = 1
    text = "Hello world python programming"

    # Index document
    search_engine.index(doc_id, text)

    # De-index
    search_engine.deindex(doc_id)

    # Verify de-indexed
    assert not search_engine.is_indexed(doc_id)

    # Search should return empty results
    tokens, results = search_engine.search("python")
    assert len(results) == 0


def test_clear_all_indexes(search_engine):
    """Test clearing all indexes"""
    # Index multiple documents
    test_docs = {1: "First document", 2: "Second document", 3: "Third document"}

    for doc_id, text in test_docs.items():
        search_engine.index(doc_id, text)

    # Clear all indexes
    search_engine.clear_all_indexes()

    # Verify all indexes cleared
    for doc_id in test_docs:
        assert not search_engine.is_indexed(doc_id)


def test_max_results(search_engine):
    """Test maximum results limit"""
    # Index multiple documents with same common keyword
    for i in range(500):
        search_engine.index(i, f"document with common word {i}")

    tokens, results = search_engine.search("common", limit=300)

    # Verify result count does not exceed max_results
    assert len(results) <= 300


def test_ranking(search_engine):
    """
    More rigorous testing of document relevance ranking
    Verify:
    1. Documents containing more query terms should rank higher
    2. Impact of term frequency on ranking
    3. Ensure ranking is meaningful
    """
    docs = {
        1: "python is a great programming language with many python features",  # High match
        2: "python programming language concepts",  # High match
        3: "another document mentioning python once",  # Low match
        4: "completely unrelated document",  # Unrelated document
    }

    for doc_id, text in docs.items():
        search_engine.index(doc_id, text)

    tokens, results = search_engine.search("python programming")

    # Verify basic search results
    assert len(results) == 3, "Search should return results"

    # Unpack results: document IDs and relevance scores
    doc_ids, scores = zip(*results)

    # Verify ranking order and relevance
    assert 1 in doc_ids, "Document containing all query terms should be in results"
    assert 3 in doc_ids, "Document partially matching query terms should be in results"

    # Check ranking order
    top_doc_id = doc_ids[0]

    assert top_doc_id in (1, 2), "Document with highest match should be first"

    # Verify score decreasing
    for i in range(1, len(scores)):
        assert scores[i - 1] >= scores[i], "Relevance scores should decrease"

    # Check reasonableness of relevance scores
    assert all(
        score > 0 for score in scores
    ), "All matching document scores should be positive"

    # Print debug information (optional)
    print("Query terms:", tokens)
    print("Document ranking:", list(zip(doc_ids, scores)))


def test_empty_query(search_engine):
    """Test empty query and no-results scenarios"""
    # Index some documents
    search_engine.index(1, "Sample document")

    # Test empty query
    tokens, results = search_engine.search("")
    assert tokens == []
    assert results == []

    # Test no matching query
    tokens, results = search_engine.search("xyzabc")
    assert tokens == ["xyzabc"]
    assert results == []


def test_special_characters(search_engine):
    """Boundary case testing"""
    # Special characters
    special_text = "!@#$%^&* special characters"
    search_engine.index(2, special_text)
    tokens, results = search_engine.search("special")
    assert len(results) > 0


def test_invalid_input(search_engine):
    """Test handling of invalid input"""
    with pytest.raises(Exception):
        search_engine.index(None, None)

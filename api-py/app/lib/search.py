import json
import math
import re
import string
from collections import Counter
from functools import reduce
from typing import Tuple
from redis import Redis

import jieba

# most common words in English and wikipedia:
# https://en.wikipedia.org/wiki/Most_common_words_in_English
PUNCTUATION_CN = '，、；：。？！‘’“”（）「」【】《》……'
HTML_TAG = re.compile(r'<[^>]*>')
PUNCTUATION = re.compile('[%s]' % re.escape(string.punctuation + PUNCTUATION_CN))

# fmt: off
STOP_WORDS = frozenset(
    ('a', 'an', 'and', 'are', 'as', 'at', 'be', 'by', 'can', 'for', 'from', 'have', 'if', 'in', 'is', 'it',
     'may', 'not', 'of', 'on', 'or', 'tbd', 'that', 'the', 'this', 'to', 'us', 'we', 'when', 'will', 'with',
     'yet', 'you', 'your', '的', '了', '和', '着', '与')
)


def html_filter(html_str: str) -> str:
    """Converts HTML to plain text."""
    return HTML_TAG.sub(' ', html_str)


def punctuation_filter(text: str) -> str:
    """Removes punctuation."""
    return PUNCTUATION.sub(' ', text)


def tokenize(text: str) -> list[str]:
    """Tokenizes text using jieba (a Chinese word segmentation library)."""
    return [token for token in jieba.cut_for_search(text) if token.strip()]


def stopword_filter(tokens: list[str]) -> list[str]:
    """Filters out stop words."""
    return [token for token in tokens if token not in STOP_WORDS]


def analyze(text: str) -> list[str]:
    """Apply text processing pipeline, and returns a clean list of tokens."""
    text = html_filter(text)
    text = punctuation_filter(text)

    tokens = tokenize(text)
    tokens = [token.lower() for token in tokens]
    tokens = stopword_filter(tokens)

    return tokens


class FullTextSearch:
    """A full-text search implementation using Redis as the backend.

    This class provides methods for indexing, searching, and managing
    documents in a Redis-based full-text search system. It supports
    efficient indexing and retrieval with TF-IDF based ranking.

    Potential Improvement: Add error handling and logging.
    """

    # Use Redis to store:
    # - Document token frequencies `doc:{id}:tokens`
    # - Total number of documents `docs:count`
    # - Reverse index of tokens to document IDs `token:{token}:docs`
    db: Redis

    def __init__(self, db: Redis, key_prefix: str = ''):
        self.db = db
        self.key_prefix = key_prefix

        assert db.connection_pool.connection_kwargs.get('decode_responses') is True

    def is_indexed(self, id: int) -> bool:
        """Check if a document has already been indexed."""
        return self.db.exists(self._doc_tokens_key(id)) == 1

    @property
    def doc_count(self) -> int:
        """Retrieve the number of documents indexed."""
        rv = self.db.get(self._doc_count_key())
        return int(str(rv)) if rv else 0

    def index(self, id: int, text: str) -> None:
        """Index a new document in the search system.

        Tokenizes the text, creates token frequency count, and updates
        Redis indexes. Skips indexing if no tokens are found.
        """

        if self.is_indexed(id):
            return self.reindex(id, text)

        tokens = analyze(text)
        if not tokens:
            return

        token_frequency = Counter(tokens)

        pipe = self.db.pipeline()
        pipe.set(self._doc_tokens_key(id), json.dumps(token_frequency))
        pipe.incr(self._doc_count_key())
        for token in set(tokens):
            pipe.sadd(self._token_docs_key(token), id)
        pipe.execute()

    def reindex(self, id: int, text: str) -> None:
        """Update the index for an existing document.

        Removes old tokens not present in the new text and adds new tokens.
        """

        if not self.is_indexed(id):
            return self.index(id, text)

        new_tokens = analyze(text)
        if not new_tokens:
            self.deindex(id)
            return

        new_token_frequency = Counter(new_tokens)
        old_token_frequency = json.loads(self.db.get(self._doc_tokens_key(id)) or "{}")
        if not old_token_frequency:
            raise ValueError('token frequency cannot be empty')

        new_token_set = set(new_token_frequency)
        old_token_set = set(old_token_frequency)

        pipe = self.db.pipeline()
        pipe.set(self._doc_tokens_key(id), json.dumps(new_token_frequency))
        for token in old_token_set - new_token_set:
            pipe.srem(self._token_docs_key(token), id)
        for token in new_token_set - old_token_set:
            pipe.sadd(self._token_docs_key(token), id)
        pipe.execute()

    def deindex(self, id: int) -> None:
        """Remove all indexes for a specific document."""
        token_frequency = json.loads(self.db.get(self._doc_tokens_key(id)) or "{}")
        if not token_frequency:
            raise ValueError('token frequency cannot be empty')

        pipe = self.db.pipeline()
        pipe.delete(self._doc_tokens_key(id))
        pipe.decr(self._doc_count_key())
        for token in set(token_frequency):
            pipe.srem(self._token_docs_key(token), id)
        pipe.execute()

    def clear_all_indexes(self) -> None:
        """Clear all indexed data from the Redis database.

        Removes all keys related to document indexing.
        """
        db = self.db
        for prefix in [f"{self.key_prefix}doc:", f"{self.key_prefix}token:"]:
            keys = db.keys(prefix + '*')
            if keys:
                db.delete(*keys)

    def _update_inverted_index(self, id: int, old_tokens: set[str], new_tokens: set[str]) -> None:
        tokens_to_del = old_tokens - new_tokens
        tokens_to_add = new_tokens - old_tokens
        pipe = self.db.pipeline()
        for token in tokens_to_del:
            pipe.srem(self._token_docs_key(token), id)
        for token in tokens_to_add:
            pipe.sadd(self._token_docs_key(token), id)
        pipe.execute()

    def search(self, query: str, partial: bool = True, limit: int = 0) -> tuple[list[str], list[Tuple[int, float]]]:
        """Perform a full-text search on the indexed documents.

        Returns:
            Tuple[List[str], List[Tuple[int, float]]]:
            - A list of processed query tokens
            - A list of (document_id, relevance_score) tuples

        Potential Optimization: Add support for partial matches or fuzzy search.
        """

        tokens = analyze(query)
        if not tokens:
            return tokens, []

        indexes = [self.db.smembers(self._token_docs_key(token)) for token in tokens]
        op = set.union if partial else set.intersection
        ids = [int(idx) for idx in reduce(op, indexes)]
        if not ids:
            return tokens, []

        ranked_results = sorted(self._rank(tokens, ids), key=lambda x: x[1], reverse=True)

        if limit and len(ids) > limit:
            ranked_results = ranked_results[:limit]

        return tokens, ranked_results

    def _rank(self, tokens: list[str], ids: list[int]) -> list[Tuple[int, float]]:
        """
        Calculate relevance scores for documents using TF-IDF.

        Args:
            tokens (List[str]): The query tokens.
            ids (List[int]): The document IDs to rank.

        Returns:
            List[Tuple[int, float]]: Ranked list of (document_id, score) pairs.
        """

        results = []

        pipe = self.db.pipeline()
        pipe.get(self._doc_count_key())
        for id in ids:
            pipe.get(self._doc_tokens_key(id))
        for token in tokens:
            pipe.scard(self._token_docs_key(token))
        rv = pipe.execute()

        total_docs = int(rv[0])
        token_frequencies = [json.loads(item) for item in rv[1 : len(ids) + 1]]
        doc_frequencies = rv[len(ids) + 1 :]

        for id, token_freq in zip(ids, token_frequencies):
            score = 0.0
            matching_terms = 0

            for token, df in zip(tokens, doc_frequencies):
                tf = int(token_freq.get(token, 0))
                if tf > 0:
                    matching_terms += 1

                # Use an improved TF calculation: 1 + log(tf) to reduce the weight of high-frequency terms
                normalized_tf = 1 + math.log10(tf) if tf > 0 else 0

                idf = math.log10(max(1, total_docs / df)) if df > 0 else 0
                score += normalized_tf * idf

            # Apply a length normalization factor to avoid advantages for long documents
            total_terms = sum(token_freq.values())
            if total_terms > 0:
                score = score / math.sqrt(total_terms)

            # Calculate query term coverage
            coverage_ratio = matching_terms / len(tokens)

            if coverage_ratio == 1.0:
                # Weight documents with full query term matches
                score *= 2.0
            else:
                score *= coverage_ratio

            results.append((id, score))
        return results

    def _doc_count_key(self) -> str:
        return f"{self.key_prefix}doc:count"

    def _doc_tokens_key(self, id: int) -> str:
        return f"{self.key_prefix}doc:{id}:tokens"

    def _token_docs_key(self, token: str) -> str:
        return f"{self.key_prefix}token:{token}:docs"

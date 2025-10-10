package fulltext

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

var tokenizer = NewGseTokenizer()

// setupTestRedis creates a test Redis client
func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use a separate test database
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Clear test database
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("Failed to flush test database: %v", err)
	}

	return client
}

// teardownTestRedis cleans up after tests
func teardownTestRedis(t *testing.T, client *redis.Client) {
	ctx := context.Background()
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Errorf("Failed to flush test database: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Errorf("Failed to close Redis client: %v", err)
	}
}

func TestGseTokenizer_Cut(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		contains []string
	}{
		{
			name:     "English text",
			input:    "The quick brown fox jumps over the lazy dog",
			wantLen:  17,
			contains: []string{"quick", "brown", "fox"},
		},
		{
			name:     "Chinese text",
			input:    "我爱自然语言处理",
			wantLen:  4,
			contains: []string{"我", "爱", "自然语言", "处理"},
		},
		{
			name:     "Mixed text",
			input:    "Python是一门编程语言",
			wantLen:  4,
			contains: []string{"python", "编程语言"},
		},
		{
			name:    "Empty text",
			input:   "",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizer.Cut(tt.input)
			if len(tokens) != tt.wantLen {
				t.Errorf("Cut() got %d tokens, want %d", len(tokens), tt.wantLen)
			}

			for _, want := range tt.contains {
				found := false
				for _, token := range tokens {
					if token == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Cut() missing expected token %q in result %v", want, tokens)
				}
			}
		})
	}
}

func TestGseTokenizer_Analyze(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    []string
		notContains []string
	}{
		{
			name:        "Remove HTML tags",
			input:       "<p>Hello <strong>world</strong></p>",
			contains:    []string{"hello", "world"},
			notContains: []string{"p", "strong"},
		},
		{
			name:        "Remove punctuation",
			input:       "Hello, world! How are you?",
			contains:    []string{"hello", "world", "how"},
			notContains: []string{",", "!", "?"},
		},
		{
			name:        "Remove stop words (English)",
			input:       "The quick brown fox is jumping",
			contains:    []string{"quick", "brown", "fox", "jumping"},
			notContains: []string{"the", "is"},
		},
		{
			name:        "Remove stop words (Chinese)",
			input:       "我爱自然语言处理和机器学习",
			contains:    []string{"爱", "自然语言", "处理", "机器", "学习"},
			notContains: []string{"和", "的"},
		},
		{
			name:        "Lowercase conversion",
			input:       "HELLO World",
			contains:    []string{"hello", "world"},
			notContains: []string{"HELLO", "World"},
		},
		{
			name:     "Mixed content",
			input:    "<h1>Python编程</h1>, 很有趣!",
			contains: []string{"python", "编程", "很", "有趣"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizer.Analyze(tt.input)

			for _, want := range tt.contains {
				found := false
				for _, token := range tokens {
					if token == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Analyze() missing expected token %q in result %v", want, tokens)
				}
			}

			for _, unwanted := range tt.notContains {
				for _, token := range tokens {
					if token == unwanted {
						t.Errorf("Analyze() contains unwanted token %q in result %v", unwanted, tokens)
					}
				}
			}
		})
	}
}

func TestFullTextSearch_IndexAndIndexed(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:")
	ctx := context.Background()

	// Test indexing English document
	t.Run("Index English document", func(t *testing.T) {
		err := fts.Index(ctx, 1, "The quick brown fox jumps over the lazy dog")
		if err != nil {
			t.Fatalf("Index() error = %v", err)
		}

		indexed, err := fts.Indexed(ctx, 1)
		if err != nil {
			t.Fatalf("Indexed() error = %v", err)
		}
		if !indexed {
			t.Error("Expected document to be indexed")
		}
	})

	// Test indexing Chinese document
	t.Run("Index Chinese document", func(t *testing.T) {
		err := fts.Index(ctx, 2, "我爱自然语言处理和机器学习")
		if err != nil {
			t.Fatalf("Index() error = %v", err)
		}

		indexed, err := fts.Indexed(ctx, 2)
		if err != nil {
			t.Fatalf("Indexed() error = %v", err)
		}
		if !indexed {
			t.Error("Expected document to be indexed")
		}
	})

	// Test not indexed
	t.Run("Check not indexed document", func(t *testing.T) {
		indexed, err := fts.Indexed(ctx, 999)
		if err != nil {
			t.Fatalf("Indexed() error = %v", err)
		}
		if indexed {
			t.Error("Expected document to not be indexed")
		}
	})

	// Test empty document
	t.Run("Index empty document", func(t *testing.T) {
		err := fts.Index(ctx, 3, "")
		if err != nil {
			t.Fatalf("Index() error = %v", err)
		}

		// Empty document should not be indexed
		indexed, err := fts.Indexed(ctx, 3)
		if err != nil {
			t.Fatalf("Indexed() error = %v", err)
		}
		if indexed {
			t.Error("Empty document should not be indexed")
		}
	})
}

func TestFullTextSearch_GetDocCount(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:")
	ctx := context.Background()

	// Initial count should be 0
	count, err := fts.GetDocCount(ctx)
	if err != nil {
		t.Fatalf("GetDocCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Expected initial doc count to be 0, got %d", count)
	}

	// Index documents
	documents := map[int64]string{
		1: "The quick brown fox",
		2: "我爱编程",
		3: "Machine learning is amazing",
	}

	for id, text := range documents {
		if err := fts.Index(ctx, id, text); err != nil {
			t.Fatalf("Failed to index document %d: %v", id, err)
		}
	}

	// Count should be 3
	count, err = fts.GetDocCount(ctx)
	if err != nil {
		t.Fatalf("GetDocCount() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Expected doc count to be 3, got %d", count)
	}
}

func TestFullTextSearch_Reindex(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:")
	ctx := context.Background()

	// Index initial document
	if err := fts.Index(ctx, 1, "The quick brown fox"); err != nil {
		t.Fatalf("Index() error = %v", err)
	}

	// Reindex with new content
	if err := fts.Reindex(ctx, 1, "The slow red turtle"); err != nil {
		t.Fatalf("Reindex() error = %v", err)
	}

	// Search for old content
	_, results, err := fts.Search(ctx, "quick brown fox")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) > 0 {
		t.Error("Should not find old content after reindex")
	}

	// Search for new content
	_, results, err = fts.Search(ctx, "slow red turtle")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 || results[0].ID != 1 {
		t.Errorf("Expected to find document 1 with new content, got %v", results)
	}

	// Test reindexing non-existent document (should work like Index)
	if err := fts.Reindex(ctx, 2, "新的中文文档"); err != nil {
		t.Fatalf("Reindex() on new document error = %v", err)
	}

	indexed, err := fts.Indexed(ctx, 2)
	if err != nil {
		t.Fatalf("Indexed() error = %v", err)
	}
	if !indexed {
		t.Error("Expected document 2 to be indexed after reindex")
	}
}

func TestFullTextSearch_Deindex(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:")
	ctx := context.Background()

	// Index document
	if err := fts.Index(ctx, 1, "The quick brown fox"); err != nil {
		t.Fatalf("Index() error = %v", err)
	}

	// Verify indexed
	indexed, err := fts.Indexed(ctx, 1)
	if err != nil {
		t.Fatalf("Indexed() error = %v", err)
	}
	if !indexed {
		t.Fatal("Document should be indexed")
	}

	// Deindex
	if err := fts.Deindex(ctx, 1); err != nil {
		t.Fatalf("Deindex() error = %v", err)
	}

	// Verify not indexed
	indexed, err = fts.Indexed(ctx, 1)
	if err != nil {
		t.Fatalf("Indexed() error = %v", err)
	}
	if indexed {
		t.Error("Document should not be indexed after deindex")
	}

	// Verify doc count decreased
	count, err := fts.GetDocCount(ctx)
	if err != nil {
		t.Fatalf("GetDocCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Expected doc count to be 0 after deindex, got %d", count)
	}

	// Search should return no results
	_, results, err := fts.Search(ctx, "quick brown")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected no search results after deindex, got %v", results)
	}
}

func TestFullTextSearch_SearchEnglish(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:")
	ctx := context.Background()

	// Index documents
	documents := map[int64]string{
		1: "The quick brown fox jumps over the lazy dog",
		2: "A fast brown fox is running in the forest",
		3: "The slow turtle walks under the bright sun",
		4: "Machine learning and artificial intelligence",
		5: "Deep learning neural networks",
	}

	for id, text := range documents {
		if err := fts.Index(ctx, id, text); err != nil {
			t.Fatalf("Failed to index document %d: %v", id, err)
		}
	}

	tests := []struct {
		name       string
		query      string
		wantIDs    []int64
		minResults int
	}{
		{
			name:       "Single word",
			query:      "fox",
			wantIDs:    []int64{1, 2},
			minResults: 2,
		},
		{
			name:       "Multiple words",
			query:      "brown fox",
			wantIDs:    []int64{1, 2},
			minResults: 2,
		},
		{
			name:       "Common word",
			query:      "learning",
			wantIDs:    []int64{4, 5},
			minResults: 2,
		},
		{
			name:       "No match",
			query:      "elephant",
			wantIDs:    []int64{},
			minResults: 0,
		},
		{
			name:       "Phrase query",
			query:      "machine learning",
			wantIDs:    []int64{4},
			minResults: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, results, err := fts.Search(ctx, tt.query)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) < tt.minResults {
				t.Errorf("Expected at least %d results, got %d", tt.minResults, len(results))
			}

			for _, wantID := range tt.wantIDs {
				found := false
				for _, result := range results {
					if result.ID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find document %d in results", wantID)
				}
			}

			// Verify scores are in descending order
			for i := 1; i < len(results); i++ {
				if results[i-1].Score < results[i].Score {
					t.Error("Results are not sorted by score in descending order")
				}
			}
		})
	}
}

func TestFullTextSearch_SearchChinese(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:")
	ctx := context.Background()

	// Index Chinese documents
	documents := map[int64]string{
		1: "我爱自然语言处理和机器学习",
		2: "深度学习是机器学习的一个分支",
		3: "自然语言处理技术发展迅速",
		4: "人工智能改变世界",
		5: "神经网络模型训练",
	}

	for id, text := range documents {
		if err := fts.Index(ctx, id, text); err != nil {
			t.Fatalf("Failed to index document %d: %v", id, err)
		}
	}

	tests := []struct {
		name       string
		query      string
		wantIDs    []int64
		minResults int
	}{
		{
			name:       "Single term",
			query:      "机器学习",
			wantIDs:    []int64{1, 2},
			minResults: 2,
		},
		{
			name:       "Multiple terms",
			query:      "自然语言处理",
			wantIDs:    []int64{1, 3},
			minResults: 2,
		},
		{
			name:       "Common term",
			query:      "学习",
			wantIDs:    []int64{1, 2},
			minResults: 2,
		},
		{
			name:       "No match",
			query:      "区块链",
			wantIDs:    []int64{},
			minResults: 0,
		},
		{
			name:       "Specific phrase",
			query:      "深度学习",
			wantIDs:    []int64{2},
			minResults: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, results, err := fts.Search(ctx, tt.query)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) < tt.minResults {
				t.Errorf("Expected at least %d results, got %d", tt.minResults, len(results))
			}

			for _, wantID := range tt.wantIDs {
				found := false
				for _, result := range results {
					if result.ID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find document %d in results", wantID)
				}
			}
		})
	}
}

func TestFullTextSearch_SearchMixed(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:")
	ctx := context.Background()

	// Index mixed language documents
	documents := map[int64]string{
		1: "Python是一门流行的编程语言",
		2: "JavaScript and TypeScript are popular",
		3: "Go语言适合构建高性能服务",
		4: "Machine learning with Python",
		5: "使用TensorFlow进行深度学习",
	}

	for id, text := range documents {
		if err := fts.Index(ctx, id, text); err != nil {
			t.Fatalf("Failed to index document %d: %v", id, err)
		}
	}

	tests := []struct {
		name       string
		query      string
		wantIDs    []int64
		minResults int
	}{
		{
			name:       "English term",
			query:      "Python",
			wantIDs:    []int64{1, 4},
			minResults: 2,
		},
		{
			name:       "Chinese term",
			query:      "编程语言",
			wantIDs:    []int64{1},
			minResults: 1,
		},
		{
			name:       "Mixed query",
			query:      "Go语言",
			wantIDs:    []int64{3},
			minResults: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, results, err := fts.Search(ctx, tt.query)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) < tt.minResults {
				t.Errorf("Expected at least %d results, got %d", tt.minResults, len(results))
			}

			for _, wantID := range tt.wantIDs {
				found := false
				for _, result := range results {
					if result.ID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find document %d in results", wantID)
				}
			}
		})
	}
}

func TestFullTextSearch_PartialMatch(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	ctx := context.Background()

	// Test with partial match enabled
	t.Run("Partial match enabled", func(t *testing.T) {
		fts := NewFullTextSearch(client, tokenizer, true, 100, "test:fts:partial:")

		documents := map[int64]string{
			1: "The quick brown fox",
			2: "The lazy dog",
			3: "The quick dog",
		}

		for id, text := range documents {
			if err := fts.Index(ctx, id, text); err != nil {
				t.Fatalf("Failed to index document %d: %v", id, err)
			}
		}

		// Search for "quick dog" - should match docs 1, 2, 3 (union)
		_, results, err := fts.Search(ctx, "quick dog")
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results with partial match, got %d", len(results))
		}

		if err := fts.ClearAllIndexes(ctx); err != nil {
			t.Fatalf("ClearAllIndexes() error = %v", err)
		}
	})

	// Test with partial match disabled
	t.Run("Partial match disabled", func(t *testing.T) {
		fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:exact:")

		documents := map[int64]string{
			1: "The quick brown fox",
			2: "The lazy dog",
			3: "The quick dog",
		}

		for id, text := range documents {
			if err := fts.Index(ctx, id, text); err != nil {
				t.Fatalf("Failed to index document %d: %v", id, err)
			}
		}

		// Search for "quick dog" - should only match doc 3 (intersection)
		_, results, err := fts.Search(ctx, "quick dog")
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if len(results) != 1 || results[0].ID != 3 {
			t.Errorf("Expected only document 3 with exact match, got %v", results)
		}

		if err := fts.ClearAllIndexes(ctx); err != nil {
			t.Fatalf("ClearAllIndexes() error = %v", err)
		}
	})
}

func TestFullTextSearch_MaxResults(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, true, 3, "test:fts:")
	ctx := context.Background()

	// Index many documents
	for i := 1; i <= 10; i++ {
		text := "machine learning and artificial intelligence"
		if err := fts.Index(ctx, int64(i), text); err != nil {
			t.Fatalf("Failed to index document %d: %v", i, err)
		}
	}

	// Search should return only maxResults
	_, results, err := fts.Search(ctx, "machine learning")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results (maxResults), got %d", len(results))
	}
}

func TestFullTextSearch_ClearAllIndexes(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:")
	ctx := context.Background()

	// Index documents
	documents := map[int64]string{
		1: "The quick brown fox",
		2: "我爱编程",
		3: "Machine learning",
	}

	for id, text := range documents {
		if err := fts.Index(ctx, id, text); err != nil {
			t.Fatalf("Failed to index document %d: %v", id, err)
		}
	}

	// Verify documents are indexed
	count, err := fts.GetDocCount(ctx)
	if err != nil {
		t.Fatalf("GetDocCount() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 documents before clear, got %d", count)
	}

	// Clear all indexes
	if err := fts.ClearAllIndexes(ctx); err != nil {
		t.Fatalf("ClearAllIndexes() error = %v", err)
	}

	// Verify all indexes are cleared
	count, err = fts.GetDocCount(ctx)
	if err != nil {
		t.Fatalf("GetDocCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 documents after clear, got %d", count)
	}

	// Search should return no results
	_, results, err := fts.Search(ctx, "fox")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected no results after clear, got %v", results)
	}
}

func TestFullTextSearch_Scoring(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:")
	ctx := context.Background()

	// Index documents with varying relevance
	documents := map[int64]string{
		1: "machine learning machine learning machine learning", // High frequency
		2: "machine learning and deep learning",                 // Multiple query terms
		3: "artificial intelligence and neural networks",        // No query terms
		4: "machine learning",                                   // Exact match
	}

	for id, text := range documents {
		if err := fts.Index(ctx, id, text); err != nil {
			t.Fatalf("Failed to index document %d: %v", id, err)
		}
	}

	_, results, err := fts.Search(ctx, "machine learning")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	// Verify results are scored
	if len(results) < 2 {
		t.Fatalf("Expected at least 2 results, got %d", len(results))
	}

	// All scores should be positive
	for _, result := range results {
		if result.Score <= 0 {
			t.Errorf("Expected positive score, got %f for doc %d", result.Score, result.ID)
		}
	}

	// Document 4 (exact match) should have high coverage boost
	foundDoc4 := false
	for _, result := range results {
		if result.ID == 4 {
			foundDoc4 = true
			if result.Score < 0.1 {
				t.Errorf("Expected higher score for exact match document, got %f", result.Score)
			}
		}
	}
	if !foundDoc4 {
		t.Error("Expected to find exact match document in results")
	}
}

func TestFullTextSearch_EmptyQuery(t *testing.T) {
	client := setupTestRedis(t)
	defer teardownTestRedis(t, client)

	fts := NewFullTextSearch(client, tokenizer, false, 100, "test:fts:")
	ctx := context.Background()

	// Index a document
	if err := fts.Index(ctx, 1, "The quick brown fox"); err != nil {
		t.Fatalf("Index() error = %v", err)
	}

	// Search with empty query
	tokens, results, err := fts.Search(ctx, "")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(tokens) != 0 {
		t.Errorf("Expected no tokens for empty query, got %v", tokens)
	}

	if len(results) != 0 {
		t.Errorf("Expected no results for empty query, got %v", results)
	}
}

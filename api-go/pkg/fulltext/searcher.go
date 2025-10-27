package fulltext

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"

	t "github.com/cymoo/pebble/pkg/util/types"
	"github.com/redis/go-redis/v9"
)

// TokenFrequency stores token frequencies for a document
type TokenFrequency map[string]int

// FullTextSearch provides full-text search functionality
type FullTextSearch struct {
	client    *redis.Client
	tokenizer Tokenizer
	keyPrefix string
}

// NewFullTextSearch creates a new FullTextSearch instance
func NewFullTextSearch(
	client *redis.Client,
	tokenizer Tokenizer,
	keyPrefix string,
) *FullTextSearch {
	return &FullTextSearch{
		client:    client,
		tokenizer: tokenizer,
		keyPrefix: keyPrefix,
	}
}

// Indexed checks if a document is indexed
func (f *FullTextSearch) Indexed(ctx context.Context, id int64) (bool, error) {
	exists, err := f.client.Exists(ctx, f.docTokensKey(id)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// GetDocCount returns the total number of indexed documents
func (f *FullTextSearch) GetDocCount(ctx context.Context) (int64, error) {
	val, err := f.client.Get(ctx, f.docCountKey()).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse doc count: %w", err)
	}
	return count, nil
}

// Index adds a document to the search index
func (f *FullTextSearch) Index(ctx context.Context, id int64, text string) error {
	indexed, err := f.Indexed(ctx, id)
	if err != nil {
		return err
	}

	if indexed {
		return f.Reindex(ctx, id, text)
	}

	// Tokenize text
	tokens := f.tokenizer.Analyze(text)
	if len(tokens) == 0 {
		return nil
	}

	// Calculate token frequencies
	tokenFreq := countFrequencies(tokens)
	freqJSON, err := json.Marshal(tokenFreq)
	if err != nil {
		return err
	}

	// Use pipeline for atomic operations
	pipe := f.client.Pipeline()
	pipe.Set(ctx, f.docTokensKey(id), freqJSON, 0)
	pipe.Incr(ctx, f.docCountKey())

	// Add document ID to token sets
	for token := range tokenFreq {
		pipe.SAdd(ctx, f.tokenDocsKey(token), id)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// Reindex updates an existing document in the index
func (f *FullTextSearch) Reindex(ctx context.Context, id int64, text string) error {
	indexed, err := f.Indexed(ctx, id)
	if err != nil {
		return err
	}

	if !indexed {
		return f.Index(ctx, id, text)
	}

	newTokens := f.tokenizer.Analyze(text)
	if len(newTokens) == 0 {
		return f.Deindex(ctx, id)
	}

	// Get old token frequencies
	var oldFreq TokenFrequency
	data, err := f.client.Get(ctx, f.docTokensKey(id)).Result()
	if err != nil {
		return fmt.Errorf("token frequency of doc %d not found: %w", id, err)
	}

	// Unmarshal old frequencies
	if err := json.Unmarshal([]byte(data), &oldFreq); err != nil {
		return err
	}

	// Calculate new frequencies
	newFreq := countFrequencies(newTokens)
	freqJSON, err := json.Marshal(newFreq)
	if err != nil {
		return err
	}

	// Calculate differences
	oldTokenSet := t.NewSet[string]()
	for token := range oldFreq {
		oldTokenSet.Add(token)
	}

	newTokenSet := t.NewSet(newTokens...)

	tokensToRemove := oldTokenSet.Difference(newTokenSet)
	tokensToAdd := newTokenSet.Difference(oldTokenSet)

	// Update index
	pipe := f.client.Pipeline()
	pipe.Set(ctx, f.docTokensKey(id), freqJSON, 0)

	// Remove document ID from old token sets and add to new token sets
	for token := range tokensToRemove {
		pipe.SRem(ctx, f.tokenDocsKey(token), id)
	}
	for token := range tokensToAdd {
		pipe.SAdd(ctx, f.tokenDocsKey(token), id)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// Deindex removes a document from the index
func (f *FullTextSearch) Deindex(ctx context.Context, id int64) error {
	var tokenFreq TokenFrequency
	data, err := f.client.Get(ctx, f.docTokensKey(id)).Result()
	if err != nil {
		return fmt.Errorf("token frequency of doc %d not found: %w", id, err)
	}

	if err := json.Unmarshal([]byte(data), &tokenFreq); err != nil {
		return err
	}

	// Remove document from index, update counts, and remove from token sets
	pipe := f.client.Pipeline()
	pipe.Del(ctx, f.docTokensKey(id))
	pipe.Decr(ctx, f.docCountKey())

	for token := range tokenFreq {
		pipe.SRem(ctx, f.tokenDocsKey(token), id)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// SearchResult represents a search result with ID and score
type SearchResult struct {
	ID    int64
	Score float64
}

// Search performs a full-text search
// query: the search query string
// partial: if true, performs a partial match (OR); if false, performs an exact match (AND)
// limit: maximum number of results to return (0 for no limit)
// Returns the tokens, ranked results, and any error encountered
func (f *FullTextSearch) Search(ctx context.Context, query string, partial bool, limit int) ([]string, []SearchResult, error) {
	tokens := f.tokenizer.Analyze(query)
	if len(tokens) == 0 {
		return tokens, []SearchResult{}, nil
	}

	// Retrieve document IDs for each token
	pipe := f.client.Pipeline()
	cmds := make([]*redis.StringSliceCmd, len(tokens))

	for i, token := range tokens {
		cmds[i] = pipe.SMembers(ctx, f.tokenDocsKey(token))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return tokens, nil, err
	}

	// Collect document IDs
	docSets := make([]map[int64]struct{}, len(cmds))
	for i, cmd := range cmds {
		members, _ := cmd.Result()
		docSets[i] = make(map[int64]struct{})
		for _, member := range members {
			if id, err := strconv.ParseInt(member, 10, 64); err == nil {
				docSets[i][id] = struct{}{}
			}
		}
	}

	// Combine document ID sets based on partial flag
	var ids map[int64]struct{}
	if partial {
		// Union
		ids = make(map[int64]struct{})
		for _, set := range docSets {
			for id := range set {
				ids[id] = struct{}{}
			}
		}
	} else {
		// Intersection
		if len(docSets) == 0 {
			return tokens, []SearchResult{}, nil
		}

		ids = docSets[0]
		for i := 1; i < len(docSets); i++ {
			newIds := make(map[int64]struct{})
			for id := range ids {
				if _, exists := docSets[i][id]; exists {
					newIds[id] = struct{}{}
				}
			}
			ids = newIds
		}
	}

	if len(ids) == 0 {
		return tokens, []SearchResult{}, nil
	}

	// Rank results
	rankedResults, err := f.rank(ctx, tokens, ids)
	if err != nil {
		return tokens, nil, err
	}

	// Sort by score descending
	sort.Slice(rankedResults, func(i, j int) bool {
		return rankedResults[i].Score > rankedResults[j].Score
	})

	// Limit results
	if limit > 0 && len(rankedResults) > limit {
		rankedResults = rankedResults[:limit]
	}

	return tokens, rankedResults, nil
}

// Rank calculates TF-IDF scores for documents
func (f *FullTextSearch) rank(ctx context.Context, tokens []string, ids map[int64]struct{}) ([]SearchResult, error) {
	totalDocs, err := f.GetDocCount(ctx)
	if err != nil {
		return nil, err
	}
	totalDocsFloat := float64(totalDocs)

	// Get token frequencies for all documents
	idList := make([]int64, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}

	// Get token frequencies for each document
	tokenFreqs := make([]TokenFrequency, len(idList))
	for i, id := range idList {
		data, err := f.client.Get(ctx, f.docTokensKey(id)).Result()
		if err != nil {
			return nil, fmt.Errorf("token frequency of doc %d not found: %w", id, err)
		}
		if err := json.Unmarshal([]byte(data), &tokenFreqs[i]); err != nil {
			return nil, err
		}
	}

	// Get document frequencies for each token
	pipe := f.client.Pipeline()
	dfCmds := make([]*redis.IntCmd, len(tokens))
	for i, token := range tokens {
		dfCmds[i] = pipe.SCard(ctx, f.tokenDocsKey(token))
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	docFreqs := make([]float64, len(tokens))
	for i, cmd := range dfCmds {
		docFreqs[i] = float64(cmd.Val())
	}

	// Calculate scores
	results := make([]SearchResult, len(idList))
	for i, id := range idList {
		tokenFreq := tokenFreqs[i]
		score := 0.0
		matchingTerms := 0

		for j, token := range tokens {
			tf := float64(tokenFreq[token])
			if tf > 0.0 {
				matchingTerms++
			}

			// Normalized TF
			normalizedTF := 0.0
			if tf > 0.0 {
				normalizedTF = 1.0 + math.Log10(tf)
			}

			// IDF
			idf := 0.0
			if docFreqs[j] > 0.0 {
				idf = math.Log10(math.Max(totalDocsFloat/docFreqs[j], 1.0))
			}

			score += normalizedTF * idf
		}

		// Length normalization
		totalTerms := 0
		for _, count := range tokenFreq {
			totalTerms += count
		}
		if totalTerms > 0 {
			score /= math.Sqrt(float64(totalTerms))
		}

		// Query term coverage
		coverageRatio := float64(matchingTerms) / float64(len(tokens))
		if coverageRatio > 0.999 {
			score *= 2.0
		} else {
			score *= coverageRatio
		}

		results[i] = SearchResult{ID: id, Score: score}
	}

	return results, nil
}

// ClearIndex removes all indexes with the configured prefix
func (f *FullTextSearch) ClearIndex(ctx context.Context) error {
	keys, err := f.client.Keys(ctx, f.keyPrefix+"*").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := f.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	return nil
}

// Key generation helpers
func (f *FullTextSearch) docCountKey() string {
	return f.keyPrefix + "count"
}

func (f *FullTextSearch) docTokensKey(id int64) string {
	return fmt.Sprintf("%s%d:tokens", f.keyPrefix, id)
}

func (f *FullTextSearch) tokenDocsKey(token string) string {
	return fmt.Sprintf("%s%s:docs", f.keyPrefix, token)
}

// Helper functions
func countFrequencies(tokens []string) map[string]int {
	freq := make(map[string]int)
	for _, token := range tokens {
		freq[token]++
	}
	return freq
}

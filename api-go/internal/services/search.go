package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/yanyiwu/gojieba"
)

var (
	punctuationRegex = regexp.MustCompile(`\p{P}`)
	htmlTagRegex     = regexp.MustCompile(`<[^>]*>`)
	stopWords        = map[string]struct{}{
		"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "by": {},
		"can": {}, "for": {}, "from": {}, "have": {}, "if": {}, "in": {}, "is": {},
		"it": {}, "may": {}, "not": {}, "of": {}, "on": {}, "or": {}, "tbd": {},
		"that": {}, "the": {}, "this": {}, "to": {}, "us": {}, "we": {}, "when": {},
		"will": {}, "with": {}, "yet": {}, "you": {}, "your": {},
		"的": {}, "了": {}, "和": {}, "着": {}, "与": {},
	}
)

// Tokenizer interface for text tokenization
type Tokenizer interface {
	Cut(text string) []string
	Analyze(text string) []string
}

// JiebaTokenizer implements Tokenizer using gojieba
type JiebaTokenizer struct {
	jieba *gojieba.Jieba
	once  sync.Once
}

// NewJiebaTokenizer creates a new JiebaTokenizer
func NewJiebaTokenizer() *JiebaTokenizer {
	return &JiebaTokenizer{
		jieba: gojieba.NewJieba(),
	}
}

// Close releases jieba resources
func (j *JiebaTokenizer) Close() {
	j.jieba.Free()
}

// Cut tokenizes text into words
func (j *JiebaTokenizer) Cut(text string) []string {
	return j.jieba.CutForSearch(text, true)
}

// Analyze performs full text analysis with preprocessing
func (j *JiebaTokenizer) Analyze(text string) []string {
	// Remove HTML tags
	text = htmlTagRegex.ReplaceAllString(text, " ")

	// Remove punctuation
	text = punctuationRegex.ReplaceAllString(text, " ")

	// Tokenize
	tokens := j.Cut(text)

	// Filter and normalize
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.ToLower(strings.TrimSpace(token))
		if token != "" {
			if _, isStopWord := stopWords[token]; !isStopWord {
				result = append(result, token)
			}
		}
	}

	return result
}

// TokenFrequency stores token frequencies for a document
type TokenFrequency struct {
	Frequencies map[string]int `json:"frequencies"`
}

// FullTextSearch provides full-text search functionality
type FullTextSearch struct {
	client       *redis.Client
	tokenizer    Tokenizer
	partialMatch bool
	maxResults   int
	keyPrefix    string
}

// NewFullTextSearch creates a new FullTextSearch instance
func NewFullTextSearch(
	client *redis.Client,
	tokenizer Tokenizer,
	partialMatch bool,
	maxResults int,
	keyPrefix string,
) *FullTextSearch {
	return &FullTextSearch{
		client:       client,
		tokenizer:    tokenizer,
		partialMatch: partialMatch,
		maxResults:   maxResults,
		keyPrefix:    keyPrefix,
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

	tokens := f.tokenizer.Analyze(text)
	if len(tokens) == 0 {
		return nil
	}

	tokenFreq := countFrequencies(tokens)
	freqJSON, err := json.Marshal(TokenFrequency{Frequencies: tokenFreq})
	if err != nil {
		return err
	}

	tokenSet := uniqueTokens(tokens)

	// Use pipeline for atomic operations
	pipe := f.client.Pipeline()
	pipe.Set(ctx, f.docTokensKey(id), freqJSON, 0)
	pipe.Incr(ctx, f.docCountKey())

	for token := range tokenSet {
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

	if err := json.Unmarshal([]byte(data), &oldFreq); err != nil {
		return err
	}

	newFreq := countFrequencies(newTokens)
	freqJSON, err := json.Marshal(TokenFrequency{Frequencies: newFreq})
	if err != nil {
		return err
	}

	// Calculate differences
	oldTokenSet := make(map[string]struct{})
	for token := range oldFreq.Frequencies {
		oldTokenSet[token] = struct{}{}
	}

	newTokenSet := uniqueTokens(newTokens)

	tokensToRemove := make([]string, 0)
	for token := range oldTokenSet {
		if _, exists := newTokenSet[token]; !exists {
			tokensToRemove = append(tokensToRemove, token)
		}
	}

	tokensToAdd := make([]string, 0)
	for token := range newTokenSet {
		if _, exists := oldTokenSet[token]; !exists {
			tokensToAdd = append(tokensToAdd, token)
		}
	}

	// Update index
	pipe := f.client.Pipeline()
	pipe.Set(ctx, f.docTokensKey(id), freqJSON, 0)

	for _, token := range tokensToRemove {
		pipe.SRem(ctx, f.tokenDocsKey(token), id)
	}

	for _, token := range tokensToAdd {
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

	pipe := f.client.Pipeline()
	pipe.Del(ctx, f.docTokensKey(id))
	pipe.Decr(ctx, f.docCountKey())

	for token := range tokenFreq.Frequencies {
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
func (f *FullTextSearch) Search(ctx context.Context, query string) ([]string, []SearchResult, error) {
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

	// Combine results based on partial match setting
	var ids map[int64]struct{}
	if f.partialMatch {
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
	if len(rankedResults) > f.maxResults {
		rankedResults = rankedResults[:f.maxResults]
	}

	return tokens, rankedResults, nil
}

// rank calculates TF-IDF scores for documents
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
		tokenFreq := &tokenFreqs[i]
		score := 0.0
		matchingTerms := 0

		for j, token := range tokens {
			tf := float64(tokenFreq.Frequencies[token])
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
		for _, count := range tokenFreq.Frequencies {
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

// ClearAllIndexes removes all indexes with the configured prefix
func (f *FullTextSearch) ClearAllIndexes(ctx context.Context) error {
	prefixes := []string{
		f.keyPrefix + "doc:",
		f.keyPrefix + "token:",
	}

	for _, prefix := range prefixes {
		keys, err := f.client.Keys(ctx, prefix+"*").Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := f.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Key generation helpers
func (f *FullTextSearch) docCountKey() string {
	return f.keyPrefix + "doc:count"
}

func (f *FullTextSearch) docTokensKey(id int64) string {
	return fmt.Sprintf("%sdoc:%d:tokens", f.keyPrefix, id)
}

func (f *FullTextSearch) tokenDocsKey(token string) string {
	return fmt.Sprintf("%stoken:%s:docs", f.keyPrefix, token)
}

// Helper functions
func countFrequencies(tokens []string) map[string]int {
	freq := make(map[string]int)
	for _, token := range tokens {
		freq[token]++
	}
	return freq
}

func uniqueTokens(tokens []string) map[string]struct{} {
	unique := make(map[string]struct{})
	for _, token := range tokens {
		unique[token] = struct{}{}
	}
	return unique
}

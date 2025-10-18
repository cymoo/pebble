package util

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// IsChineseCharacter checks if a rune is a Chinese character
func IsChineseCharacter(c rune) bool {
	return c >= '\u4e00' && c <= '\u9fff'
}

// Highlight highlights all occurrences of the given tokens in the HTML string
// by wrapping them with <mark> tags, while preserving existing HTML tags.
func Highlight(html string, tokens []string) string {
	if len(tokens) == 0 {
		return html
	}

	// Sort tokens by length in descending order to match longer tokens first
	sortedTokens := make([]string, len(tokens))
	copy(sortedTokens, tokens)
	sort.Slice(sortedTokens, func(i, j int) bool {
		return len(sortedTokens[i]) > len(sortedTokens[j])
	})

	// Build regex pattern to match any of the tokens
	patterns := make([]string, len(sortedTokens))
	for i, token := range sortedTokens {
		patterns[i] = tokenToPattern(token)
	}

	// Combine HTML tag pattern with token patterns
	pattern := `(<[^>]*>)|(` + strings.Join(patterns, "|") + `)`
	re, err := regexp.Compile(pattern)
	if err != nil {
		// In case of regex compilation error, return original HTML
		return html
	}

	// Replace matches with highlighted versions, preserving HTML tags
	result := re.ReplaceAllStringFunc(html, func(match string) string {
		// If it's an HTML tag, return as is
		if len(match) > 0 && match[0] == '<' && match[len(match)-1] == '>' {
			return match
		}
		// Otherwise, highlight the matched token
		return "<mark>" + match + "</mark>"
	})

	return result
}

// tokenToPattern converts a token into a regex pattern for matching
func tokenToPattern(token string) string {
	// Escape special regex characters in the token
	escaped := regexp.QuoteMeta(token)

	// Check if the token contains any Chinese characters
	hasChinese := false
	for _, c := range token {
		if IsChineseCharacter(c) {
			hasChinese = true
			break
		}
	}

	if hasChinese {
		return escaped
	}

	// Check if it's purely alphanumeric or spaces
	isPureAlphanumeric := true
	for _, c := range token {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsSpace(c) {
			isPureAlphanumeric = false
			break
		}
	}

	if isPureAlphanumeric {
		return `\b` + escaped + `\b`
	}

	return escaped
}

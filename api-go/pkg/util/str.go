package util

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// isChineseCharacter 检查字符是否为中文字符
func IsChineseCharacter(c rune) bool {
	return c >= '\u4e00' && c <= '\u9fff'
}

// highlight 在HTML文本中标记所有出现的令牌
func Highlight(html string, tokens []string) string {
	if len(tokens) == 0 {
		return html
	}

	// 按长度降序排序令牌
	sortedTokens := make([]string, len(tokens))
	copy(sortedTokens, tokens)
	sort.Slice(sortedTokens, func(i, j int) bool {
		return len(sortedTokens[i]) > len(sortedTokens[j])
	})

	// 构建模式
	patterns := make([]string, len(sortedTokens))
	for i, token := range sortedTokens {
		patterns[i] = tokenToPattern(token)
	}

	// 组合HTML标签和令牌模式
	pattern := `(<[^>]*>)|(` + strings.Join(patterns, "|") + `)`
	re, err := regexp.Compile(pattern)
	if err != nil {
		// 如果正则编译失败，返回原始HTML
		return html
	}

	// 处理文本
	result := re.ReplaceAllStringFunc(html, func(match string) string {
		// 检查是否是HTML标签
		if len(match) > 0 && match[0] == '<' && match[len(match)-1] == '>' {
			return match // HTML标签保持不变
		}
		// 令牌匹配 - 用mark标签包裹
		return "<mark>" + match + "</mark>"
	})

	return result
}

// tokenToPattern 将令牌转换为正则表达式模式
func tokenToPattern(token string) string {
	// 对正则特殊字符进行转义
	escaped := regexp.QuoteMeta(token)

	// 检查是否包含中文字符
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

	// 检查是否全是字母数字或空格
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

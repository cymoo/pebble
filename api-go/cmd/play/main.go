package main

import (
	"fmt"

	"github.com/cymoo/pebble/pkg/fulltext"
)

func main() {
	tokenizer := fulltext.NewGseTokenizer()
	texts := []string{
		"我爱自然语言处理",
		"Python是一门编程语言",
		"我爱自然语言处理和机器学习",
		"The quick brown fox jumps over the lazy dog",
	}
	for _, text := range texts {
		tokens := tokenizer.Analyze(text)
		fmt.Printf("Text: %s; Tokens: %#v; len: %d\n", text, tokens, len(tokens))
	}
}

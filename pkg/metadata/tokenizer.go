/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata

import "strings"

var (
	// DefaultPerWordTokenizer splits the input on whitespace and filters out any 0 or 1 length words along with the common words.
	DefaultPerWordTokenizer = NewStandardTokenizer()

	// DefaultExactMatchTokenizer converts the given input to lowercase but does not do any breaks.
	DefaultExactMatchTokenizer = &exactMatchTokenizer{}

	// DefaultNopTokenizer does not tokenize making the field unsearchable, useful when not indexing for the field is required
	DefaultNopTokenizer = &nopTokenizer{}
)

// Tokenizer converts the given input into searchable tokens
type Tokenizer interface {
	Tokenize(input string) []string
}

type exactMatchTokenizer struct {
}

func (e exactMatchTokenizer) Tokenize(input string) []string {
	return []string{strings.ToLower(input)}
}

type nopTokenizer struct {
}

func (nopTokenizer) Tokenize(input string) []string {
	return nil
}

type TokenizerFunc func(string) []string

func (f TokenizerFunc) Tokenize(input string) []string {
	return f(input)
}

// TokenizerChain stitches together multitple tokenizers together.
func TokenizerChain(tokenizers ...Tokenizer) Tokenizer {
	return TokenizerFunc(func(input string) []string {
		var result []string
		for _, tokenizer := range tokenizers {
			result = append(result, tokenizer.Tokenize(input)...)
		}
		return result
	})
}

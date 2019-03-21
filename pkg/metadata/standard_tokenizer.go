/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata

import "strings"

var (
	defaultStopWords = []string{
		"and",
		"is",
		"an",
		"then",
		"the",
		"not",
		"when",
		"or",
		"to",
		"from",
		"for",
		"of",
		"if",
		"at",
		"about",
		"use",
		"with",
		"inc",
		"llc",
	}
)

type StandardTokenizerOption func(st *StandardTokenizer) bool

type SplitterFunc func(string) []string

type TrimmerFunc func(string) string

func defaultSplitter(input string) []string {
	return strings.Fields(input)
}

func defaultTrimmer(v string) string {
	return strings.Trim(v, ",:;!%$#()*\"")
}

func makeTrimmerFunc(cutset string) TrimmerFunc {
	return TrimmerFunc(func(v string) string {
		return strings.Trim(v, cutset)
	})
}

// StandardTokenizer breaks the given input into separate terms based on the supplied argument and filters any stop words before
// returning the result.
type StandardTokenizer struct {
	stopWords    map[string]bool
	splitterFunc SplitterFunc
	trimmerFunc  TrimmerFunc
}

func toSet(list []string) map[string]bool {
	set := map[string]bool{}

	for _, item := range list {
		set[item] = true
	}
	return set
}

func NewStandardTokenizer(options ...StandardTokenizerOption) Tokenizer {

	stdTokenizer := &StandardTokenizer{toSet(defaultStopWords), defaultSplitter, defaultTrimmer}

	for _, option := range options {
		option(stdTokenizer)
	}
	return stdTokenizer
}

func WithStopWords(stopWords []string) StandardTokenizerOption {
	return func(st *StandardTokenizer) bool {
		set := toSet(stopWords)
		st.stopWords = set
		return true
	}
}

func WithSplitter(separator string) StandardTokenizerOption {

	fn := defaultSplitter

	if separator != "" {
		fn = SplitterFunc(func(input string) []string {
			return strings.Split(input, separator)
		})
	}
	return func(st *StandardTokenizer) bool {
		st.splitterFunc = fn
		return true
	}
}

func WithTrimmer(cutset string) StandardTokenizerOption {
	fn := defaultTrimmer

	fn = TrimmerFunc(func(v string) string {
		return strings.Trim(v, cutset)
	})

	return func(st *StandardTokenizer) bool {
		st.trimmerFunc = fn
		return true
	}
}

func (st *StandardTokenizer) Tokenize(input string) []string {
	input = strings.ToLower(input)

	var (
		tokens []string
		terms  []string
	)

	tokens = st.splitterFunc(input)
	for _, v := range tokens {
		v = st.trimmerFunc(v)

		if len(v) == 0 || len(v) == 1 || st.stopWords[v] {
			continue
		}
		terms = append(terms, v)
	}
	return terms
}

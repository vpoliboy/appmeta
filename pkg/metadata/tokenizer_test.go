/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestExactMatchTokenizer_Tokenize(t *testing.T) {
	url := "http://Github.com/vpoliboy"

	terms := DefaultExactMatchTokenizer.Tokenize(url)
	assert.Len(t, terms, 1, "Exact match should have only one entry")
	assert.Contains(t, terms, strings.ToLower(url), "expecting some in term list")
}

func TestStandardTokenizer_TokenizeSingleline(t *testing.T) {
	description := "Some application content, and description"

	terms := DefaultPerWordTokenizer.Tokenize(description)

	assert.Equal(t, 4, len(terms), "expecting 4 terms ")
	assert.Contains(t, terms, "some", "expecting some in term list")
	assert.Contains(t, terms, "application", "expecting application in term list")
	assert.Contains(t, terms, "content", "expecting content in term list")
	assert.NotContains(t, terms, "and", "not expecting and in term list")
}

func TestStandardTokenizer_TokenizeMultiline(t *testing.T) {
	description :=
		`Some application content, and description
		  with a multiline
		`

	terms := DefaultPerWordTokenizer.Tokenize(description)

	assert.Equal(t, 5, len(terms), "expecting 4 terms ")
	assert.Contains(t, terms, "some", "expecting some in term list")
	assert.Contains(t, terms, "application", "expecting application in term list")
	assert.Contains(t, terms, "content", "expecting content in term list")
	assert.NotContains(t, terms, "with", "not expecting with in term list")
	assert.NotContains(t, terms, "and", "not expecting and in term list")
	assert.Contains(t, terms, "multiline", "expecting multiline in term list")
}

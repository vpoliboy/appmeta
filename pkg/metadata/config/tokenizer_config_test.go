/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMakeStandardTokenizerFromConfig(t *testing.T) {

	config :=
		`{
  "tokenizerConfig" : [
    {
      "name": "SpaceDelimitedWordTokenizer",
      "type" : "Standard",
      "config": {
        "stopWords": [
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
          "llc"
        ],
        "cutset": " \\n\\t,:;!%$#()\\*",
        "separator": ""
      }
    },
    {
      "name": "ExactWordTokenizer",
      "type": "ExactMatch"
    },
    {
      "name": "ChainedTokenizer",
      "type": "Chain",
      "config": {
        "tokenizers" : ["SpaceDelimitedWordTokenizer", "ExactWordTokenizer"]
      }
    }
  ],

  "fieldConfig": {
    "name": "ChainedTokenizer",
    "email": "ExactWordTokenizer",
    "title": "ExactWordTokenizer",
    "version": "ExactWordTokenizer",
    "company": "SpaceDelimitedWordTokenizer",
    "website": "ExactWordTokenizer",
    "source": "ExactWordTokenizer",
    "license": "ExactWordTokenizer",
    "description": "SpaceDelimitedWordTokenizer"
  }
}
`
	a := &AnalyzerConfig{}
	err := json.Unmarshal([]byte(config), a)
	assert.Nil(t, err)

	tokenizersMapping, err := CreateFieldTokenizers(a)
	assert.Nil(t, err)
	assert.True(t, len(tokenizersMapping) > 0)

	for k, v := range tokenizersMapping {
		switch k {
		case "name": // Chained
			tokens := v.Tokenize("Vijay Poliboyina")
			assert.Len(t, tokens, 3)
			t.Log(tokens)
		case "email": // Exact
			tokens := v.Tokenize("vijaykp@gmail.com")
			assert.Len(t, tokens, 1)
			t.Log(tokens)
		case "title": // ExactWordTokenizer
			tokens := v.Tokenize("appmeta")
			assert.Len(t, tokens, 1)
			t.Log(tokens)
		case "version": // "ExactWordTokenizer"
			tokens := v.Tokenize("0.1.1")
			assert.Len(t, tokens, 1)
			t.Log(tokens)
		case "company": // "SpaceDelimitedWordTokenizer"
			tokens := v.Tokenize("Feye Inc")
			// inc is filtered out
			assert.Len(t, tokens, 1)
			t.Log(tokens)
		case "website": // "ExactWordTokenizer",
			tokens := v.Tokenize("https://fireeye.com")
			assert.Len(t, tokens, 1)
			t.Log(tokens)
		case "source": // "ExactWordTokenizer",
			tokens := v.Tokenize("https://github.com/fireeye")
			assert.Len(t, tokens, 1)
			t.Log(tokens)
		case "license": // "ExactWordTokenizer",
			tokens := v.Tokenize("Apache-2.0")
			assert.Len(t, tokens, 1)
			t.Log(tokens)
		case "description": // "SpaceDelimitedWordTokenizer"
			tokens := v.Tokenize("awesome app from feye")
			// from is filtered out
			assert.Len(t, tokens, 3)
			t.Log(tokens)
		}
	}

}

package config

import (
	"encoding/json"
	"errors"
	"github.com/vpoliboy/appmeta/pkg/metadata"
	"strings"
)

var (
	errInvalidStdTokenizerConfig   = errors.New("invalid standard tokenizer config")
	errInvalidChainTokenizerConfig = errors.New("invalid chain tokenizer config")
	errFieldTokenizerConfig        = errors.New("invalid field to tokenizer config")
)

type AnalyzerConfig struct {
	// Tokenizer Id to Config map
	TokenizerConfigs []struct {
		Name   string          `json:"name"`
		Type   string          `json:"type"`
		Config json.RawMessage `json:"config,omitempty"`
	} `json:"tokenizerConfig"`

	// Field to Tokenizer map
	FieldConfig map[metadata.SearchField]string `json:"fieldConfig"`
}

type StandardTokenizerConfig struct {
	StopWords []string `json:"stopWords"`
	Separator string   `json:"separator,omitempty"`
	Cutset    string   `json:"cutset"`
}

type ChainedTokenizerConfig struct {
	TokenizerNames []string `json:"tokenizers"`
}

func CreateFieldTokenizers(config *AnalyzerConfig) (map[metadata.SearchField]metadata.Tokenizer, error) {

	tokenizers := map[string]metadata.Tokenizer{}

	for _, v := range config.TokenizerConfigs {
		switch strings.ToLower(v.Type) {
		case "exactmatch":
			tokenizers[v.Name] = metadata.DefaultExactMatchTokenizer
		case "nop":
			tokenizers[v.Name] = metadata.DefaultNopTokenizer
		case "standard":
			tokenizer, err := MakeStandardTokenizerFromConfig(v.Config)
			if err != nil {
				return nil, err
			}
			tokenizers[v.Name] = tokenizer
		case "chain":
			tokenizer, err := MakeTokenizerChain(v.Config, tokenizers)
			if err != nil {
				return nil, err
			}
			tokenizers[v.Name] = tokenizer
		}
	}

	fieldTokenizerMapping := map[metadata.SearchField]metadata.Tokenizer{}
	for k, v := range config.FieldConfig {
		if tokenizer, ok := tokenizers[v]; ok {
			fieldTokenizerMapping[k] = tokenizer
			continue
		}
		return nil, errFieldTokenizerConfig
	}
	return fieldTokenizerMapping, nil
}

func MakeStandardTokenizerFromConfig(jsonConfig json.RawMessage) (metadata.Tokenizer, error) {
	config := &StandardTokenizerConfig{}
	if json.Unmarshal(jsonConfig, config) != nil {
		return nil, errInvalidStdTokenizerConfig
	}
	return metadata.NewStandardTokenizer(
		metadata.WithStopWords(config.StopWords),
		metadata.WithSplitter(config.Separator),
		metadata.WithTrimmer(config.Cutset)), nil
}

func MakeTokenizerChain(jsonConfig json.RawMessage, tokenizers map[string]metadata.Tokenizer) (metadata.Tokenizer, error) {
	config := &ChainedTokenizerConfig{}
	if json.Unmarshal(jsonConfig, config) != nil {
		return nil, errInvalidChainTokenizerConfig
	}

	chain := make([]metadata.Tokenizer, 0, len(config.TokenizerNames))

	for _, tokenizerName := range config.TokenizerNames {
		if tokenizer, ok := tokenizers[tokenizerName]; ok {
			chain = append(chain, tokenizer)
			continue
		}
		return nil, errInvalidChainTokenizerConfig
	}
	return metadata.TokenizerChain(chain...), nil
}

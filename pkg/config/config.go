/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package config

import (
	"encoding/json"
	"errors"
	"github.com/vpoliboy/appmeta/pkg/metadata"
	mconfig "github.com/vpoliboy/appmeta/pkg/metadata/config"
	"os"
	"path/filepath"
)

const (
	analyzerJson = "analyzer.json"
)

var (
	ErrNoAnalyzerFileExists  = errors.New("missing analyzer file")
	ErrInvalidAnalyzerFormat = errors.New("invalid analyzer file format")
)

func LoadAnalyzerConfig(confDir string) (map[metadata.SearchField]metadata.Tokenizer, error) {

	fileLocation := filepath.Join(confDir, analyzerJson)
	if _, err := os.Stat(fileLocation); os.IsNotExist(err) {
		return nil, ErrNoAnalyzerFileExists
	}

	f, err := os.Open(fileLocation)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	analyzerConfig := &mconfig.AnalyzerConfig{}
	err = json.NewDecoder(f).Decode(analyzerConfig)
	if err != nil {
		return nil, ErrInvalidAnalyzerFormat
	}
	return mconfig.CreateFieldTokenizers(analyzerConfig);
}

/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata

var (
	defaultSearchFieldTokenizerMapping = map[SearchField]Tokenizer{
		// title and version fields are exact match search
		titleField:   DefaultExactMatchTokenizer,
		versionField: DefaultExactMatchTokenizer,

		// company fields and split around white spaces into tokens.
		companyField: DefaultPerWordTokenizer,

		// website and source (both URLs) are exact match fields
		websiteField: DefaultExactMatchTokenizer,
		sourceField:  DefaultExactMatchTokenizer,

		// license is also exact match assuming they its an Identifier rather than the text
		licenseField: DefaultExactMatchTokenizer,

		// description is full text so word tokenizer.
		descriptionField: DefaultPerWordTokenizer,

		// Name is a special field that is both exactmatch and tokenized for searching on both first and last names.
		nameField: TokenizerChain(DefaultPerWordTokenizer, DefaultExactMatchTokenizer),

		// Email is exact match field
		emailField: DefaultExactMatchTokenizer,
	}
)

type Analyzer struct {
	tokenizerMapping map[SearchField]Tokenizer
}

// Analyzer breaks down the given metadata into a data structure thats used for indexing the fields.
// For a given Metadata struct, Analyzer will
//    1. break down each individual field into a list of tokens/terms based on the tokenizer configured for that field
//    2. Creates the mapping from the fieldname to the list of tokens/terms for that field.
func (a *Analyzer) AnalyzePayload(p *Metadata) map[SearchField][]string {

	// TODO: make the tokenizer configurable.
	tokens := map[SearchField][]string{
		// title and version fields are exact match search
		titleField:   a.tokenizerMapping[titleField].Tokenize(p.Title),
		versionField: a.tokenizerMapping[versionField].Tokenize(p.Version),

		// company fields and split around white spaces into tokens.
		companyField: a.tokenizerMapping[companyField].Tokenize(p.Company),

		// website and source (both URLs) are exact match fields
		websiteField: a.tokenizerMapping[websiteField].Tokenize(p.Website),
		sourceField:  a.tokenizerMapping[sourceField].Tokenize(p.SourceURL),

		// license is also exact match assuming they its an Identifier rather than the text
		licenseField: a.tokenizerMapping[licenseField].Tokenize(p.License),

		// description is full text so word tokenizer.
		descriptionField: a.tokenizerMapping[descriptionField].Tokenize(p.Description),
	}
	for _, m := range p.Maintainers {
		for k, v := range a.analyzeMaintainer(m) {
			tokens[k] = append(tokens[k], v...)
		}
	}
	return tokens
}

func (a *Analyzer) analyzeMaintainer(m Maintainer) map[SearchField][]string {
	return map[SearchField][]string{
		// Name is a special field that is both exactmatch and tokenized for searching on both first and last names.
		nameField: a.tokenizerMapping[nameField].Tokenize(m.Name),

		// Email is exact match field
		emailField:  a.tokenizerMapping[emailField].Tokenize(m.Email),
	}
}

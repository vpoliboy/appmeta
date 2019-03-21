/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemoryIndexer_IndexAndSearch(t *testing.T) {

	indexer := newInMemoryIndexer(logrus.New())
	analyzer := &Analyzer{defaultSearchFieldTokenizerMapping}

	hits, err := indexer.Search(Query{nameField: "vijay"})
	assert.Nil(t, err)
	assert.Len(t, hits, 0)


	m := &Metadata{
		Title:   "appmeta",
		Version: "0.1.0",
		Maintainers: []Maintainer{
			{"Vijay Poliboyina", "vijaykp@gmail.com"},
		},
		Company:     "feye Inc.",
		Website:     "https://feye.io",
		SourceURL:   "https://github.com/feye.io",
		License:     "Apache-2.0",
		Description: "App metadata service",
	}
	searchFields := analyzer.AnalyzePayload(m)

	_, err = indexer.Index(searchFields, m)


	m2 := &Metadata{
		Title:   "appmeta2",
		Version: "0.1.0",
		Maintainers: []Maintainer{
			{"V Poliboyina", "vijaykp@gmail.com"},
		},
		Company:     "feye Inc.",
		Website:     "https://feye.io",
		SourceURL:   "https://github.com/feye.io",
		License:     "Apache-2.0",
		Description: "App metadata service",
	}
	searchFields2 := analyzer.AnalyzePayload(m2)

	_, err = indexer.Index(searchFields2, m2)

	assert.Nil(t, err)

	hits, err = indexer.Search(Query{nameField: "vijay"})
	assert.Nil(t, err)
	assert.Len(t, hits, 1)

	hits, err = indexer.Search(Query{descriptionField: "metadata"})
	assert.Nil(t, err)
	assert.Len(t, hits, 2)

	hits, err = indexer.Search(Query{nameField: "poliboyina"})
	assert.Nil(t, err)
	assert.Len(t, hits, 2)

	hits, err = indexer.Search(Query{nameField: "poliboyina", titleField: "appmeta"})
	assert.Nil(t, err)
	assert.Len(t, hits, 1)

	hits, err = indexer.Search(Query{nameField: "v poliboyina"})
	assert.Nil(t, err)
	assert.Len(t, hits, 1)

	hits, err = indexer.Search(Query{companyField: "feye"})
	assert.Nil(t, err)
	assert.Len(t, hits, 2)

	hits, err = indexer.Search(Query{companyField: "cfeye"})
	assert.Nil(t, err)
	assert.Len(t, hits, 0)

}

func TestInMemoryIndexer_ConcurrentIndexAndSearch(t *testing.T) {
	t.SkipNow()

	count := 1000 * 4

	indexer := newInMemoryIndexer(logrus.New())
	analyzer := &Analyzer{defaultSearchFieldTokenizerMapping}


	for i := 0; i < count; i++ {
		t.Run(fmt.Sprintf("p-%d", i), func(tt *testing.T) {
			tt.Parallel()


			m := &Metadata{
				Title:   fmt.Sprintf("appmeta%d", i),
				Version: "0.1.0",
				Maintainers: []Maintainer{
					{"V Poliboyina", "vijaykp@gmail.com"},
				},
				Company:     "feye Inc.",
				Website:     "https://feye.io",
				SourceURL:   "https://github.com/feye.io",
				License:     "Apache-2.0",
				Description: fmt.Sprintf("App metadata service%d", i),
			}
			s := analyzer.AnalyzePayload(m)
			_, err := indexer.Index(s, m)
			assert.Nil(tt, err)

			hits, err := indexer.Search(Query{titleField: fmt.Sprintf("appmeta%d", i)})
			assert.Nil(tt, err)
			assert.Len(tt, hits, 1)
		})
	}
	assert.Equal(t, count, int(indexer.Size()))

}

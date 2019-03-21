/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata

import (
	"context"
	"fmt"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"strings"
)

type Service interface {
	Search(context.Context, Query) ([]*MetadataWithID, error)
	GetAll(context.Context) ([]*MetadataWithID, error)
	Delete(context.Context, uuid.UUID) bool
	Get(context.Context, uuid.UUID) (*MetadataWithID, error)
	Insert(*Metadata) (uuid.UUID, error)
	Version() string
	Health() error
	Shutdown(context.Context) error
}

type metadataSearchService struct {
	indexer Indexer
	analyzer *Analyzer
	logger  *logrus.Logger
}

type ServiceOption func(*metadataSearchService) bool

func NewService(logger *logrus.Logger, opts ...ServiceOption) Service {
	s := &metadataSearchService{
		indexer: newInMemoryIndexer(logger),
		analyzer: &Analyzer{defaultSearchFieldTokenizerMapping},
		logger:  logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithMappings(tokenizerMapping map[SearchField]Tokenizer) ServiceOption {
	return ServiceOption(func(s *metadataSearchService) bool {
		s.analyzer = &Analyzer{tokenizerMapping}
		return true
	})
}

func (svc *metadataSearchService) processQuery(_ context.Context, query Query) (Query, error) {
	processedQuery := Query{}
	for k, v := range query {
		if _, ok := allowedSearchFields[k]; !ok {
			return nil, validation.NewInternalError(fmt.Errorf(" %s is not a valid search field", k))
		}
		processedQuery[k] = strings.ToLower(v)
	}
	return processedQuery, processedQuery.Validate()
}

func (svc *metadataSearchService) Search(ctx context.Context, query Query) ([]*MetadataWithID, error) {
	var err error
	if query, err = svc.processQuery(ctx, query); err != nil {
		return nil, err
	}
	return svc.indexer.Search(query)
}

func (svc *metadataSearchService) GetAll(_ context.Context) ([]*MetadataWithID, error) {
	return svc.indexer.GetAll()
}

func (svc *metadataSearchService) Insert(payload *Metadata) (uuid.UUID, error) {
	var err error
	if err = payload.Validate(); err != nil {
		return uuid.Nil, err
	}

	// breakdown the Metadata into fields to tokens maps
	searchTerms := svc.analyzer.AnalyzePayload(payload)
	svc.logger.Debug("Metadata Tokens: ", searchTerms)
	return svc.indexer.Index(searchTerms, payload)
}

func (svc *metadataSearchService) Delete(_ context.Context, id uuid.UUID) bool {
	return svc.indexer.Delete(id) == nil
}

func (svc *metadataSearchService) Get(_ context.Context, id uuid.UUID) (*MetadataWithID, error) {
	return svc.indexer.Get(id)
}

func (svc *metadataSearchService) Shutdown(_ context.Context) error {
	return nil
}

func (svc *metadataSearchService) Version() string {
	return "0.1.0"
}

func (svc *metadataSearchService) Health() error {
	return svc.indexer.Health()
}

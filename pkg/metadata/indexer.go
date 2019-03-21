/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata

import (
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
)

var (
	errUUIDGenError = errors.New("error generating uuid")
	errNotFound     = errors.New("not found")
)

var (
	noHits = []*MetadataWithID{}
)

// MetadataWithID is the structure that is stored in the indexer
type MetadataWithID struct {

	// Auto generated on a new indexing request
	ID uuid.UUID `json:"_id" yaml:"_id"`

	// User-supplied metadata structure
	*Metadata
}

// Interface for the implementation of the Indexer
type Indexer interface {

	// Indexes the fields for searchability and stores in the repo, returns an auto-generated UUID on success.
	Index(map[SearchField][]string, *Metadata) (uuid.UUID, error)

	// Delete and remove the indexing structures corresponding to the metadata ID from the repo
	Delete(uuid.UUID) error

	// Singlefield search - Returns all the metadata payloads that match the 'value' for the given 'field'
	SearchBySingleField(field SearchField, value string) ([]*MetadataWithID, error)

	// Multifield search - Returns all the metadata payloads that match ALL the values for the given fields
	// This is AND filter in that all the filters have to match for the metadata to be considered a hit
	Search(Query) ([]*MetadataWithID, error)

	// Returns all the metadatas
	GetAll() ([]*MetadataWithID, error)

	// Get the metadata with the specified ID, if no ID is there then errNotFound is returned
	Get(uuid.UUID) (*MetadataWithID, error)

	// Health API
	Health() error

	// Number of metadata payloads that are currently indexed.
	Size() uint64
}

type uuidSet map[uuid.UUID]bool
type TermIndex map[string]uuidSet

// inMemoryIndexer implements the indexer interface by two data structures
//	1. searchIndex of type map[SearchField]map[string]UUID
//        - maintains the inverted index of fields -> fieldValues/terms -> metadata UUIDs
//  2. uuid2MetadataIndex of type similar to ConcurrentMap[uuid.UUID]Metadata
//        - maintains the UUID to metadata payload mapping
type inMemoryIndexer struct {
	searchMutex *sync.RWMutex
	searchIndex map[SearchField]TermIndex

	// similar to ConcurrentMap[uuid.UUID]Metadata
	uuid2MetadataIndex *sync.Map

	// atomic counter for number of items in the index
	metadataCount uint64
	logger        *logrus.Logger
}

func newInMemoryIndexer(logger *logrus.Logger) Indexer {
	return &inMemoryIndexer{
		searchMutex:        &sync.RWMutex{},
		searchIndex:        map[SearchField]TermIndex{},
		uuid2MetadataIndex: &sync.Map{},
		metadataCount:      0,
		logger:             logger,
	}
}

func (repo *inMemoryIndexer) Index(searchTerms map[SearchField][]string, p *Metadata) (uuid.UUID, error) {

	var (
		termValueIndex TermIndex
		ok             bool
	)

	metadataID, err := uuid.NewUUID()
	if err != nil {
		return uuid.Nil, errUUIDGenError
	}

	// From this point it is safe to assume no errors or inconsistencies will happen.
	repo.uuid2MetadataIndex.Store(metadataID, p)
	atomic.AddUint64(&repo.metadataCount, 1)

	// Modify the search inverted index as the Metadata is already inserted into the uuidset
	repo.searchMutex.Lock()
	defer repo.searchMutex.Unlock()

	for fieldName, terms := range searchTerms {

		// Check for the existence of the field key
		if termValueIndex, ok = repo.searchIndex[fieldName]; !ok {
			termValueIndex = TermIndex{}
			repo.searchIndex[fieldName] = termValueIndex
		}

		// Modify the terms to metadata id mapping
		for _, term := range terms {
			uuids, ok := termValueIndex[term]
			if !ok {
				uuids = uuidSet{}
				termValueIndex[term] = uuids
			}
			uuids[metadataID] = true
		}
	}
	return metadataID, nil
}

func (repo *inMemoryIndexer) Delete(uuid.UUID) error {
	return errors.New("delete not implemented")
}

func (repo *inMemoryIndexer) SearchBySingleField(fieldName SearchField, term string) ([]*MetadataWithID, error) {
	if fieldName == anyField {
		return repo.SearchAny(term)
	}

	repo.searchMutex.RLock()
	defer repo.searchMutex.RUnlock()

	uuids, err := repo.getUUIDsByField(fieldName, term)
	if err != nil || len(uuids) == 0 {
		return nil, err
	}

	// We got the UUIDs of all matched Metadata structures, now get the actual values.
	metadatas := make([]*MetadataWithID, 0, len(uuids))
	for uuid := range uuids {
		if v, ok := repo.uuid2MetadataIndex.Load(uuid); ok {
			metadatas = append(metadatas, &MetadataWithID{uuid, v.(*Metadata)})
		}
	}
	return metadatas, nil
}

func (repo *inMemoryIndexer) getUUIDsByField(fieldName SearchField, term string) (uuidSet, error) {
	var (
		termIndex TermIndex
		ok        bool
	)

	if termIndex, ok = repo.searchIndex[fieldName]; !ok {
		return nil, nil
	}
	return termIndex[term], nil
}

// SearchAny tries to match the given term against all the fields. Just a wrapper
// around the getUUIDsAnyField with the rw mutex.
func (repo *inMemoryIndexer) SearchAny(term string) ([]*MetadataWithID, error) {
	repo.searchMutex.RLock()
	defer repo.searchMutex.RUnlock()

	metadataIds, err := repo.getUUIDsAnyField(term)
	if err != nil {
		return nil, err
	}
	return repo.get(metadataIds)
}

func (repo *inMemoryIndexer) merge(first, second uuidSet) uuidSet {
	union := uuidSet{}
	for k := range first {
		union[k] = true
	}
	for k := range second {
		union[k] = true
	}
	return union
}

// Loops through all the fields trying to search for the given term in the corresponding index/map.
// Slightly expensive which can be avoided with another dedicated index but with more memory requirements.
func (repo *inMemoryIndexer) getUUIDsAnyField(term string) (uuidSet, error) {
	uuids := uuidSet{}
	for _, termIndex := range repo.searchIndex {
		termUUIds := termIndex[term]
		uuids = repo.merge(uuids, termUUIds)
	}
	return uuids, nil
}

// Searches the given filters/query against the inverted index and returns the hits.
// Each filter/field->term expression is evaluated separately and a matched Metadata is checked against a previously
// matched set to be considered as a hit. Therefore a metadata item is considered a hit only if it matches against all
// the filters specified in the query. Any is a special meta search field that is used to match against all the other
// field values.
func (repo *inMemoryIndexer) Search(query Query) ([]*MetadataWithID, error) {
	var (
		filteredUUIDs uuidSet
		err           error
		firstTime     = true
	)

	repo.searchMutex.RLock()
	defer repo.searchMutex.RUnlock()

	// Prioritize the any field first if in the query
	if term, ok := query[anyField]; ok {
		if filteredUUIDs, err = repo.getUUIDsAnyField(term); err != nil {
			return nil, err
		}
		delete(query, anyField)
		firstTime = false
	}

	for fieldName, term := range query {
		matchedUUIDs, err := repo.getUUIDsByField(fieldName, term)
		if err != nil {
			return nil, err
		}
		if firstTime {
			filteredUUIDs = matchedUUIDs
			firstTime = false
		} else {
			filteredUUIDs = repo.intersectionOf(matchedUUIDs, filteredUUIDs)
		}
		if len(filteredUUIDs) == 0 {
			return noHits, nil
		}
	}
	return repo.get(filteredUUIDs)
}

func (repo *inMemoryIndexer) intersectionOf(first, second uuidSet) uuidSet {
	if len(first) > len(second) {
		return repo.intersectionOf(second, first)
	}

	intersection := uuidSet{}
	for k := range first {
		if _, ok := second[k]; ok {
			intersection[k] = true
		}
	}
	return intersection
}

func (repo *inMemoryIndexer) get(matchSet uuidSet) ([]*MetadataWithID, error) {
	payloads := make([]*MetadataWithID, 0, len(matchSet))
	for uuid := range matchSet {
		if v, ok := repo.uuid2MetadataIndex.Load(uuid); ok {
			payloads = append(payloads, &MetadataWithID{uuid, v.(*Metadata)})
		}
	}
	return payloads, nil
}

func (repo *inMemoryIndexer) GetAll() ([]*MetadataWithID, error) {

	payloads := make([]*MetadataWithID, 0, repo.Size()+64)
	repo.uuid2MetadataIndex.Range(func(key, val interface{}) bool {
		payloads = append(payloads, &MetadataWithID{key.(uuid.UUID), val.(*Metadata)})
		return true
	})
	return payloads, nil
}

func (repo *inMemoryIndexer) Get(id uuid.UUID) (*MetadataWithID, error) {

	if v, ok := repo.uuid2MetadataIndex.Load(id); ok {
		return &MetadataWithID{id, v.(*Metadata)}, nil
	}
	return nil, errNotFound
}

func (repo *inMemoryIndexer) Health() error {
	return nil
}

func (repo *inMemoryIndexer) Size() uint64 {
	return atomic.LoadUint64(&repo.metadataCount)
}

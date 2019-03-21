/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
)

// SearchField corresponds to the fields that are searchable in the index which are essentially all the fields
// in the metadata structure
type SearchField string

var (
	nameField        = SearchField("name")
	emailField       = SearchField("email")
	titleField       = SearchField("title")
	versionField     = SearchField("version")
	companyField     = SearchField("company")
	websiteField     = SearchField("website")
	sourceField      = SearchField("source")
	licenseField     = SearchField("license")
	descriptionField = SearchField("description")

	// Special meta search field that is used to match against the values of all of the above= fields.
	anyField         = SearchField("any")

	allowedSearchFields = map[SearchField]bool{
		nameField:        true,
		emailField:       true,
		titleField:       true,
		versionField:     true,
		companyField:     true,
		websiteField:     true,
		sourceField:      true,
		licenseField:     true,
		descriptionField: true,
		anyField:         true,
	}
)

// Query is an alias of searchfield->term map.
type Query map[SearchField]string


func (q Query) Validate() error {
	for k := range q {
		if _, ok := allowedSearchFields[k]; !ok {
			return validation.NewInternalError(fmt.Errorf(" %s is not a valid search field", k))
		}
	}
	return nil
}

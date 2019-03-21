/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestMaintainer_Validate(t *testing.T) {

	testCases := map[string]struct {
		maintainer           Maintainer
		expectedErrorMessage string
	}{
		"validMaintainer":        {maintainer: Maintainer{"Vijay Poliboyina", "vijaykp@gmail.com"}, expectedErrorMessage: ""},
		"validMaintainer2":       {maintainer: Maintainer{"Vijay K Poliboyina", "vijaykp@gmail.com"}, expectedErrorMessage: ""},
		"invalidMaintainerEmail": {maintainer: Maintainer{"Vijay Poliboyina", "vijaykpgmail.com"}, expectedErrorMessage: errMessageInvalidEmail},
		"invalidMaintainerName":  {maintainer: Maintainer{"Vijay+Poliboyina", "vijaykp@gmail.com"}, expectedErrorMessage: errMessageInvalidNameFormat},
	}

	for k, v := range testCases {
		t.Run(k, func(tt *testing.T) {
			err := v.maintainer.Validate()
			t.Log(err)
			if v.expectedErrorMessage == "" {
				assert.Nil(tt, err)
			} else {
				assert.True(tt, strings.Contains(err.Error(), v.expectedErrorMessage))
			}
		})
	}
}

func Testmetadata_Validate(t *testing.T) {

	testCases := map[string]struct {
		metadata             Metadata
		expectedErrorMessage string
	}{
		"valid": {
			metadata: Metadata{
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
			},
			expectedErrorMessage: "",
		},
		"missingVersion": {
			metadata: Metadata{
				Title: "App w/ missing version",
				Maintainers: []Maintainer{
					{"Vijay Poliboyina", "vijaykp@gmail.com"},
					{"V Poliboy", "vijaykp@hotmail.com"},
				},
				Company:     "xCompany Inc.",
				Website:     "https://website.com",
				SourceURL:   "https://github.com/company.repo",
				License:     "Apache-2.0",
				Description: "some markdown",
			},
			expectedErrorMessage: "Version:",
		},
		"invalidEmail": {
			metadata: Metadata{
				Title:   "App w/ missing version",
				Version: "1.0.0",
				Maintainers: []Maintainer{
					{"Vijay Poliboyina", "vijaykp@gmail.com"},
					{"V Poliboy", "vijaykphotmail.com"},
				},
				Company:     "xCompany Inc.",
				Website:     "https://website.com",
				SourceURL:   "https://github.com/company.repo",
				License:     "Apache-2.0",
				Description: "some markdown",
			},
			expectedErrorMessage: "Email:",
		},
		"invalidMaintainerCount": {
			metadata: Metadata{
				Title:       "App w/ missing version",
				Version:     "1.0.0",
				Maintainers: []Maintainer{},
				Company:     "xCompany Inc.",
				Website:     "https://website.com",
				SourceURL:   "https://github.com/company.repo",
				License:     "Apache-2.0",
				Description: "some markdown",
			},
			expectedErrorMessage: "Maintainers:",
		},
	}

	for k, v := range testCases {
		t.Run(k, func(tt *testing.T) {
			err := v.metadata.Validate()
			if v.expectedErrorMessage == "" {
				assert.Nil(tt, err)
			} else {
				if assert.NotNil(tt, err) {
					t.Log(err, v.expectedErrorMessage)
					assert.True(tt, strings.Contains(err.Error(), v.expectedErrorMessage))
				}
			}
		})
	}
}

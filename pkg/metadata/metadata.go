/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package metadata

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"regexp"
)

const (
	nameRegexString = "(.*)\\s(.*)"
)

const (
	errMessageInvalidEmail      = "invalid email format"
	errMessageInvalidNameFormat = "invalid name, must match regex: " + nameRegexString

	errMessageInvalidVersion         = "not in SemVer format"
	errMessageInvalidURLFormat       = "not an URL"
	errMessageInvalidMaintainerCount = "must have atleast one maintainer with not more than 1024"
	errMessageInvalidLength          = "length must be between 4 and 64 characters"
	errMessageInvalidLengthLong      = "length must be between 4 and 1024 characters"
)

var (
	nameRegexp, _ = regexp.Compile(nameRegexString)
)

type Maintainer struct {
	Name  string `json:"name" yaml:"name"`
	Email string `json:"email" yaml:"email"`
}

type Metadata struct {
	Title       string       `json:"title" yaml:"title"`
	Version     string       `json:"version" yaml:"version"`
	Maintainers []Maintainer `json:"maintainers" yaml:"maintainers"`
	Company     string       `json:"company" yaml:"company"`
	Website     string       `json:"website" yaml:"website"`
	SourceURL   string       `json:"source" yaml:"source"`
	License     string       `json:"license" yaml:"license"`
	Description string       `json:"description" yaml:"description"`
}

func (m Maintainer) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required,
			validation.RuneLength(4, 64).Error(errMessageInvalidLength),
			validation.Match(nameRegexp).Error(errMessageInvalidNameFormat)),
		validation.Field(&m.Email, validation.Required, is.Email.Error(errMessageInvalidEmail)),
	)
}

func (p Metadata) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Title, validation.Required, validation.Length(4, 64).Error(errMessageInvalidLength)),
		validation.Field(&p.Version, validation.Required, is.Semver.Error(errMessageInvalidVersion)),
		validation.Field(&p.Maintainers, validation.Required, validation.Length(1, 1024).Error(errMessageInvalidMaintainerCount)),
		validation.Field(&p.Company, validation.Required, validation.Length(4, 64).Error(errMessageInvalidLengthLong)),
		validation.Field(&p.Website, validation.Required, is.URL.Error(errMessageInvalidURLFormat)),
		validation.Field(&p.SourceURL, validation.Required, is.URL.Error(errMessageInvalidURLFormat)),
		validation.Field(&p.Description, validation.Required, validation.Length(4, 1024).Error(errMessageInvalidLengthLong)),
		validation.Field(&p.License, validation.Required, validation.Length(4, 64).Error(errMessageInvalidLength)),
	)
}

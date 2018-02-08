// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "regexp"

// GetOrganization returns the base Organization itself.
func (d *Organization) GetOrganization() *Organization {
	return d
}

var (
	// FieldPathOrganizationName is the field path for the organization name field.
	FieldPathOrganizationName = regexp.MustCompile(`^name$`)

	// FieldPathOrganizationDescription is the field path for the organization description field.
	FieldPathOrganizationDescription = regexp.MustCompile(`^description$`)

	// FieldPathOrganizationURL is the field path for the organization URL field.
	FieldPathOrganizationURL = regexp.MustCompile(`^url$`)

	// FieldPathOrganizationLocation is the field path for the organization location field.
	FieldPathOrganizationLocation = regexp.MustCompile(`^location$`)

	// FieldPathOrganizationEmail is the field path for the organization email field.
	FieldPathOrganizationEmail = regexp.MustCompile(`^email$`)
)

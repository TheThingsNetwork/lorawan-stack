// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "regexp"

// GetApplication returns the base Application itself.
func (d *Application) GetApplication() *Application {
	return d
}

var (
	// FieldPathApplicationDescription is the field path for the application description field.
	FieldPathApplicationDescription = regexp.MustCompile(`^description$`)
)

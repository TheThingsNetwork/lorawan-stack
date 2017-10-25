// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

// GetApplication returns the base Application itself.
func (d *Application) GetApplication() *Application {
	return d
}

// SetAPIKeys sets a list of APIKeys into the Application.
func (d *Application) SetAPIKeys(keys []APIKey) {
	d.APIKeys = keys
}

const (
	// These are the valid FieldMask path values for the `update_mask` in
	// the UpdateApplicationRequest message.

	// FieldPathApplicationDescription is the path value for the `description` field.
	FieldPathApplicationDescription = "description"
)

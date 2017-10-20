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
	// Valid FieldMask path values for the `update_mask` in UpdateApplicationRequest message.

	// PathApplicationDescription is the path value for the `description` field.
	PathApplicationDescription = "description"
)

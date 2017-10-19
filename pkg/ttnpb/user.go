// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

// GetUser returns the base User itself.
func (u *User) GetUser() *User {
	return u
}

const (
	// Valid FieldMask path values for the `update_mask` in UpdateUser method.

	// PathUserName is the path value for the `name` field.
	PathUserName = "name"

	// PathUserEmail is the path value for the 'email' field.
	PathUserEmail = "email"
)

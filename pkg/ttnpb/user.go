// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

// GetUser returns the base User itself.
func (u *User) GetUser() *User {
	return u
}

const (
	// These are the valid FieldMask path values for the `update_mask` in
	// the UpdateUserRequest message.

	// FieldPathUserName is the path value for the `name` field.
	FieldPathUserName = "name"

	// FieldPathUserEmail is the path value for the `email field.
	FieldPathUserEmail = "email"
)

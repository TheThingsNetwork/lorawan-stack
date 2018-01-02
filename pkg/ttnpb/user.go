// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "regexp"

// GetUser returns the base User itself.
func (u *User) GetUser() *User {
	return u
}

var (
	// FieldPathUserName is the field path for the user name field.
	FieldPathUserName = regexp.MustCompile(`^name$`)

	// FieldPathUserEmail is the field path for the user email field.
	FieldPathUserEmail = regexp.MustCompile(`^email$`)
)

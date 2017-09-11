// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

// Collaborator is an User that has rights to a certain thing.
type Collaborator struct {
	// Username is the username of the user.
	Username string `json:"username"`

	// Rights is the list of rights that the user has.
	Rights []Right `json:"rights"`
}

// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import "time"

// DefaultUser represents a user on the account server
type DefaultUser struct {
	// Username is a human-readable unique user handle
	Username string `db:"username"`

	// Email is the users primary email address
	Email string `db:"email"`

	// Validated denotes wether or not the user has validated his email address
	Validated bool `db:"validated"`

	// Password is the users password (hashed)
	Password string `json:"-" db:"password"`

	// Joined is the date the user joined
	Joined time.Time `db:"joined"`

	// Archived is the time the user archived his account
	Archived *time.Time `db:"archived"`

	// Admin denotes wether or not the user has administrative rights
	Admin bool `db:"admin"`

	// God denotes wether or not the user can enter god mode
	God bool `db:"god"`
}

// User is the interface of types that are a user.
// This can be used to build richer user types that can still be
// read and written to a database.
type User interface {
	GetUser() *DefaultUser
}

// GetUser implements User
func (u *DefaultUser) GetUser() *DefaultUser {
	return u
}

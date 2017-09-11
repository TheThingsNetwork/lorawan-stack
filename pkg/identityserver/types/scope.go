// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"database/sql/driver"
	"strings"
)

// Scope is the type that represents a scope that a client has access to.
type Scope string

const (
	// ApplicationScope represents a scope that has access to manage users applications.
	ApplicationScope Scope = "applications"

	// ProfileScope represents a scope that has r-w access to users profile.
	ProfileScope Scope = "profile"
)

// String implements the fmt.Stringer interface.
func (s Scope) String() string {
	return string(s)
}

// Scopes represents what scopes a client has access to.
type Scopes struct {
	// Application denotes whether the client has application access scope.
	Application bool

	// Profile denotes whether the client has profile access scope.
	Profile bool
}

// Value implements sql.Valuer interface.
func (s Scopes) Value() (driver.Value, error) {
	scopes := make([]string, 0)

	if s.Application {
		scopes = append(scopes, ApplicationScope.String())
	}

	if s.Profile {
		scopes = append(scopes, ProfileScope.String())
	}

	return strings.Join(scopes, ","), nil
}

// Scan implements sql.Scanner interface.
func (s *Scopes) Scan(src interface{}) error {
	scopes := strings.Split(src.(string), ",")

	for _, scope := range scopes {
		switch Scope(scope) {
		case ApplicationScope:
			s.Application = true
		case ProfileScope:
			s.Profile = true
		}
	}

	return nil
}

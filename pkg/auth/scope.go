// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

// Scope is the type of authorization scopes.
type Scope string

// String implements fmt.Stringer.
func (s Scope) String() string {
	return string(s)
}

// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"strings"

	"github.com/pkg/errors"
)

// String implements fmt.Stringer interface.
func (t ScopeType) String() string {
	typ, exists := ScopeType_name[int32(t)]
	if exists {
		return normalizeScopeType(typ)
	}

	return TYPE_INVALID.String()
}

// MarshalText implements encoding.TextMarshaler interface.
func (t ScopeType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (t *ScopeType) UnmarshalText(text []byte) error {
	val, exists := ScopeType_value[denormalizeScopeType(string(text))]
	if !exists {
		return errors.Errorf("Could not parse right `%s`", string(text))
	}

	*t = ScopeType(val)

	return nil
}

func normalizeScopeType(str string) string {
	return strings.ToLower(strings.TrimPrefix(str, "TYPE_"))
}

func denormalizeScopeType(str string) string {
	return "TYPE_" + strings.ToUpper(str)
}

// Username returns the username of the user this scope is for, or the empty string if it is not for a user.
func (s Scope) Username() string {
	if s.Type == TYPE_USER {
		return s.ID
	}
	return ""
}

// ApplicationID returns the application ID of the application this scope is for, or the empty string if it is not for an application.
func (s Scope) ApplicationID() string {
	if s.Type == TYPE_APPLICATION {
		return s.ID
	}
	return ""
}

// GatewayID returns the gateway ID of the gateway this scope is for, or the empty string if it is not for a gateway.
func (s Scope) GatewayID() string {
	if s.Type == TYPE_GATEWAY {
		return s.ID
	}
	return ""
}

// hasRight checks wether or not the right is included in this scope.
func (s Scope) hasRight(right Right) bool {
	for _, r := range s.Rights {
		if r == right {
			return true
		}
	}
	return false
}

// HasRights checks wether or not the provided right is included in the scope. It will only return true if all the provided rights are
// included in the token..
func (s Scope) HasRights(rights ...Right) bool {
	ok := true
	for _, right := range rights {
		ok = ok && s.hasRight(right)
	}

	return ok
}

func UserScope(username string, rights ...Right) Scope {
	return Scope{
		Type:   TYPE_USER,
		ID:     username,
		Rights: rights,
	}
}

func ApplicationScope(applicationID string, rights ...Right) Scope {
	return Scope{
		Type:   TYPE_APPLICATION,
		ID:     applicationID,
		Rights: rights,
	}
}

func GatewayScope(gatewayID string, rights ...Right) Scope {
	return Scope{
		Type:   TYPE_GATEWAY,
		ID:     gatewayID,
		Rights: rights,
	}
}

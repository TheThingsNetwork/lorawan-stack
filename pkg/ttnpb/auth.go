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

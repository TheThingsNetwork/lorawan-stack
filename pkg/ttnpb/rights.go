// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"strconv"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// String implements fmt.Stringer interface.
func (r Right) String() string {
	right, exists := Right_name[int32(r)]
	if exists {
		return normalizeRight(right)
	}

	return strconv.Itoa(int(r))
}

// MarshalText implements encoding.TextMarshaler interface.
func (r Right) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (r *Right) UnmarshalText(text []byte) error {
	val, exists := Right_value[denormalizeRight(string(text))]
	if !exists {
		return errors.Errorf("Could not parse right `%s`", string(text))
	}

	*r = Right(val)

	return nil
}

func normalizeRight(right string) string {
	return strings.ToLower(strings.Replace(strings.TrimLeft(right, "RIGHT_"), "_", ":", -1))
}

func denormalizeRight(right string) string {
	return "RIGHT_" + strings.ToUpper(strings.Replace(right, ":", "_", -1))
}

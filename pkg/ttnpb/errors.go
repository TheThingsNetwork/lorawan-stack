// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrEmptyUpdateMask.Register()
	ErrInvalidPathUpdateMask.Register()
}

// ErrEmptyUpdateMask is returned when the update mask is specified but empty.
var ErrEmptyUpdateMask = &errors.ErrDescriptor{
	MessageFormat: "update_mask must be non-empty",
	Code:          1,
	Type:          errors.InvalidArgument,
}

// ErrInvalidPathUpdateMask is returned when the update mask includes a wrong field path.
var ErrInvalidPathUpdateMask = &errors.ErrDescriptor{
	MessageFormat: "Invalid update_mask: `{path}` is not a valid path",
	Code:          2,
	Type:          errors.InvalidArgument,
}

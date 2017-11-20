// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrEmptyUpdateMask.Register()
}

// ErrEmptyUpdateMask is returned when the update mask is specified but empty.
var ErrEmptyUpdateMask = &errors.ErrDescriptor{
	MessageFormat: "update_mask must be non-empty",
	Code:          1,
	Type:          errors.InvalidArgument,
}

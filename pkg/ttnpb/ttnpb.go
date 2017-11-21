// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrInvalidPathUpdateMask.Register()
	ErrUpdateMaskNotFound.Register()
}

// ErrInvalidPathUpdateMask is returned when including in the `update_mask` a
// path that does not exist or is not allowed.
var ErrInvalidPathUpdateMask = &errors.ErrDescriptor{
	MessageFormat: "Invalid update_mask: `{path}` is not a valid path",
	Code:          1,
	Type:          errors.InvalidArgument,
	SafeAttributes: []string{
		"path",
	},
}

// ErrUpdateMaskNotFound is returned on update operations where `update_mask`
// field was not specfied or it was empty.
var ErrUpdateMaskNotFound = &errors.ErrDescriptor{
	MessageFormat: "update_mask must be specified and non-empty",
	Code:          2,
	Type:          errors.NotFound,
}

// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrInvalidData.Register()
	ErrEmptyFilter.Register()
	ErrNilKey.Register()
}

// ErrInvalidData represents an error returned, when data specified is not valid.
var ErrInvalidData = &errors.ErrDescriptor{
	MessageFormat: "Invalid data",
	Type:          errors.InvalidArgument,
	Code:          1,
}

// ErrEmptyFilter represents an error returned, when filter specified is empty.
var ErrEmptyFilter = &errors.ErrDescriptor{
	MessageFormat: "Filter is empty",
	Type:          errors.InvalidArgument,
	Code:          2,
}

// ErrNilKey represents an error returned, when key specified is nil.
var ErrNilKey = &errors.ErrDescriptor{
	MessageFormat: "Nil key specified",
	Type:          errors.InvalidArgument,
	Code:          3,
}

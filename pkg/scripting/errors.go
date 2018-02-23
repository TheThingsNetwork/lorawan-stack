// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package scripting

import "github.com/TheThingsNetwork/ttn/pkg/errors"

// ErrRuntime represents the ErrDescriptor of the error returned when
// there is a runtime error.
var ErrRuntime = &errors.ErrDescriptor{
	MessageFormat: "Runtime error",
	Type:          errors.External,
	Code:          1,
}

// ErrInvalidOutputType represents the ErrDescriptor of the error returned when
// the output of the script does not have a valid type.
var ErrInvalidOutputType = &errors.ErrDescriptor{
	MessageFormat:  "Invalid script output of type `{type}`",
	Type:           errors.External,
	Code:           2,
	SafeAttributes: []string{"type"},
}

// ErrInvalidOutputRange represents the ErrDescriptor of the error returned when
// the output of the script does not have a valid output range.
var ErrInvalidOutputRange = &errors.ErrDescriptor{
	MessageFormat:  "Value `{value}` does not fall in range of `{low}` to `{high}`",
	Type:           errors.External,
	Code:           3,
	SafeAttributes: []string{"low", "high", "value"},
}

// ErrInvalidInputType represents the ErrDescriptor of the error returned when
// the input for the script does not have a valid type.
var ErrInvalidInputType = &errors.ErrDescriptor{
	MessageFormat: "Invalid script input type",
	Type:          errors.InvalidArgument,
	Code:          4,
}

func init() {
	ErrRuntime.Register()
	ErrInvalidOutputType.Register()
	ErrInvalidOutputRange.Register()
	ErrInvalidInputType.Register()
}

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

// ErrInvalidOutput represents the ErrDescriptor of the error returned when
// the output of the script does not have a valid output.
var ErrInvalidOutput = &errors.ErrDescriptor{
	MessageFormat: "Invalid script output",
	Type:          errors.External,
	Code:          2,
}

func init() {
	ErrRuntime.Register()
	ErrInvalidOutput.Register()
}

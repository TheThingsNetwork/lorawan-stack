// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package scripting

import "github.com/TheThingsNetwork/ttn/pkg/errors"

// ErrRuntime represents the ErrDescriptor of the error returned when there is a runtime
// error.
var ErrRuntime = &errors.ErrDescriptor{
	MessageFormat: "Runtime error",
	Type:          errors.External,
	Code:          1,
}

func init() {
	ErrRuntime.Register()
}

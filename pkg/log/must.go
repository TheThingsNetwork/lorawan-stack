// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

// Must returns the logger that is passed or panics if it is nil
func Must(logger Interface) Interface {
	if logger != nil {
		return logger
	}

	panic("No logger attached to the context")
}

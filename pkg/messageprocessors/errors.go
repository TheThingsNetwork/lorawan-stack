// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package messageprocessors

import "github.com/TheThingsNetwork/ttn/pkg/errors"

// ErrNoMACPayload represents the ErrDescriptor of the error returned when the message does not
// contain MACPayload.
var ErrNoMACPayload = &errors.ErrDescriptor{
	MessageFormat: "Message does not contain MAC payload",
	Type:          errors.InvalidArgument,
	Code:          1,
}

// ErrInvalidInput represents the ErrDescriptor of the error returned when
// the input is not valid.
var ErrInvalidInput = &errors.ErrDescriptor{
	MessageFormat: "Invalid input",
	Type:          errors.InvalidArgument,
	Code:          2,
}

// ErrInvalidOutput represents the ErrDescriptor of the error returned when
// the output is invalid.
var ErrInvalidOutput = &errors.ErrDescriptor{
	MessageFormat: "Invalid output",
	Type:          errors.External,
	Code:          3,
}

// ErrInvalidOutputType represents the ErrDescriptor of the error returned when
// the output is not of the valid type.
var ErrInvalidOutputType = &errors.ErrDescriptor{
	MessageFormat:  "Invalid output of type `{type}`",
	Type:           errors.External,
	Code:           4,
	SafeAttributes: []string{"type"},
}

// ErrInvalidOutputRange represents the ErrDescriptor of the error returned when
// the output does not fall in a valid range.
var ErrInvalidOutputRange = &errors.ErrDescriptor{
	MessageFormat:  "Value `{value}` does not fall in range of `{low}` to `{high}`",
	Type:           errors.External,
	Code:           5,
	SafeAttributes: []string{"low", "high", "value"},
}

func init() {
	ErrNoMACPayload.Register()
	ErrInvalidInput.Register()
	ErrInvalidOutput.Register()
	ErrInvalidOutputType.Register()
	ErrInvalidOutputRange.Register()
}

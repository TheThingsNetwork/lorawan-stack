// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package networkserver

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrCorruptRegistry.Register()
	ErrFCntTooHigh.Register()
	ErrFCntTooLow.Register()
	ErrMICComputeFailed.Register()
	ErrMissingFNwkSIntKey.Register()
	ErrMissingSNwkSIntKey.Register()
	ErrNotFound.Register()
	ErrUnmarshalFailed.Register()
	ErrUnsupportedLoRaWANMajorVersion.Register()
}

// ErrUnsupportedLoRaWANMajorVersion represents error ocurring when unsupported LoRaWAN MAC version is specified.
var ErrUnsupportedLoRaWANMajorVersion = &errors.ErrDescriptor{
	MessageFormat:  "Unsupported LoRaWAN major version: `{major}`",
	Type:           errors.InvalidArgument,
	Code:           1,
	SafeAttributes: []string{"major"},
}

// ErrMissingPayload represents error ocurring when message payload is missing.
var ErrMissingPayload = &errors.ErrDescriptor{
	MessageFormat: "Message payload is missing",
	Type:          errors.InvalidArgument,
	Code:          2,
}

// ErrUnmarshalFailed represents error ocurring when payload unmarshaling fails.
var ErrUnmarshalFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to unmarshal payload",
	Type:          errors.InvalidArgument,
	Code:          3,
}

// ErrFCntTooLow represents error ocurring when FCnt is too low.
var ErrFCntTooLow = &errors.ErrDescriptor{
	MessageFormat: "FCnt is too low",
	Type:          errors.InvalidArgument,
	Code:          4,
}

// ErrFCntTooHigh represents error ocurring when FCnt is too high.
var ErrFCntTooHigh = &errors.ErrDescriptor{
	MessageFormat: "FCnt is too high",
	Type:          errors.InvalidArgument,
	Code:          5,
}

// ErrCorruptRegistry represents error ocurring when network server registry is corrupt.
var ErrCorruptRegistry = &errors.ErrDescriptor{
	MessageFormat: "Registry is corrupt",
	Type:          errors.Internal,
	Code:          6,
}

// ErrMICComputeFailed represents error ocurring when MIC computation fails.
var ErrMICComputeFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to compute MIC",
	Type:          errors.InvalidArgument,
	Code:          7,
}

// ErrNotFound represents error ocurring when device is not found.
var ErrNotFound = &errors.ErrDescriptor{
	MessageFormat: "Device not found",
	Type:          errors.NotFound,
	Code:          8,
}

// ErrMissingFNwkSIntKey represents error ocurring when FNwkSIntKey is missing.
var ErrMissingFNwkSIntKey = &errors.ErrDescriptor{
	MessageFormat: "FNwkSIntKey is unknown",
	Type:          errors.NotFound,
	Code:          9,
}

// ErrMissingSNwkSIntKey represents error ocurring when SNwkSIntKey is missing.
var ErrMissingSNwkSIntKey = &errors.ErrDescriptor{
	MessageFormat: "SNwkSIntKey is unknown",
	Type:          errors.NotFound,
	Code:          10,
}

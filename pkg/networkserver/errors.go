// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package networkserver

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrUnsupportedLoRaWANMajorVersion.Register()
	ErrMissingPayload.Register()
	ErrUnmarshalFailed.Register()
	ErrWrongPayloadType.Register()
	ErrFCntTooLow.Register()
	ErrFCntTooHigh.Register()
	ErrCorruptRegistry.Register()
}

// ErrUnsupportedLoRaWANMajorVersion represents error ocurring when unsupported LoRaWAN MAC version is specified.
var ErrUnsupportedLoRaWANMajorVersion = &errors.ErrDescriptor{
	MessageFormat: "Unsupported LoRaWAN major version: {major}",
	Type:          errors.InvalidArgument,
	Code:          1,
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

// ErrWrongPayloadType represents error ocurring when wrong payload type is received.
var ErrWrongPayloadType = &errors.ErrDescriptor{
	MessageFormat:  "Wrong payload type: {type}",
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"type"},
	Code:           4,
}

// ErrMissingDevAddr represents error ocurring when DevAddr is missing.
var ErrMissingDevAddr = &errors.ErrDescriptor{
	MessageFormat: "DevAddr is missing",
	Type:          errors.InvalidArgument,
	Code:          5,
}

// ErrFCntTooLow represents error ocurring when FCnt is too low.
var ErrFCntTooLow = &errors.ErrDescriptor{
	MessageFormat: "FCnt is too low",
	Type:          errors.InvalidArgument,
	Code:          6,
}

// ErrFCntTooHigh represents error ocurring when FCnt is too high.
var ErrFCntTooHigh = &errors.ErrDescriptor{
	MessageFormat: "FCnt is too high",
	Type:          errors.InvalidArgument,
	Code:          7,
}

// ErrCorruptRegistry represents error ocurring when network server registry is corrupt.
var ErrCorruptRegistry = &errors.ErrDescriptor{
	MessageFormat: "Registry is corrupt",
	Type:          errors.Internal,
	Code:          8,
}

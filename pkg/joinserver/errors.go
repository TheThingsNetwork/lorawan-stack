// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package joinserver

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrDeviceNotFound.Register()
	ErrTooManyDevices.Register()
	ErrWrongPayloadType.Register()
	ErrMissingPayload.Register()
	ErrWrongMICLength.Register()
	ErrInvalidMIC.Register()
	ErrAppKeyNotFound.Register()
	ErrNwkKeyNotFound.Register()
	ErrAppKeyEnvelopeNotFound.Register()
	ErrNwkKeyEnvelopeNotFound.Register()
	ErrMICCheckFailed.Register()
	ErrUnsupportedLoRaWANVersion.Register()
	ErrMissingDevAddr.Register()
	ErrMissingJoinEUI.Register()
	ErrMissingDevEUI.Register()
	ErrMissingJoinRequest.Register()
	ErrUnmarshalFailed.Register()
	ErrForwardJoinRequest.Register()
	ErrEncodeMHDRFailed.Register()
	ErrEncodePayloadFailed.Register()
	ErrComputeJoinAcceptMIC.Register()
	ErrEncryptPayloadFailed.Register()
	ErrDevNonceTooSmall.Register()
	ErrDevNonceReused.Register()
}

// ErrDeviceNotFound represents error ocurring when a device is not found.
var ErrDeviceNotFound = &errors.ErrDescriptor{
	MessageFormat: "Device not found",
	Type:          errors.NotFound,
	Code:          1,
}

// ErrTooManyDevices represents error ocurring when too many devices are associated with identifiers specified.
var ErrTooManyDevices = &errors.ErrDescriptor{
	MessageFormat: "Too many devices are associated with identifiers specified",
	Type:          errors.Conflict,
	Code:          2,
}

// ErrWrongPayloadType represents error ocurring when wrong payload type is received.
var ErrWrongPayloadType = &errors.ErrDescriptor{
	MessageFormat: "Wrong payload type: expected {expected_value}, got {got_value}",
	Type:          errors.InvalidArgument,
	Code:          3,
}

// ErrMissingPayload represents error ocurring when join request payload is missing.
var ErrMissingPayload = &errors.ErrDescriptor{
	MessageFormat: "Join request payload is missing",
	Type:          errors.InvalidArgument,
	Code:          4,
}

// ErrWrongMICLength represents error ocurring when wrong MIC has wrong length.
var ErrWrongMICLength = &errors.ErrDescriptor{
	MessageFormat: "Wrong MIC length: expected 4, got {got_value}",
	Type:          errors.InvalidArgument,
	Code:          5,
}

// ErrInvalidMIC represents error ocurring when MIC mismatch.
var ErrInvalidMIC = &errors.ErrDescriptor{
	MessageFormat: "MIC mismatch",
	Type:          errors.InvalidArgument,
	Code:          6,
}

// ErrAppKeyNotFound represents error ocurring when AppKey was not found for device.
var ErrAppKeyNotFound = &errors.ErrDescriptor{
	MessageFormat: "AppKey not found for device",
	Type:          errors.NotFound,
	Code:          7,
}

// ErrNwkKeyNotFound represents error ocurring when NwkKey was not found for device.
var ErrNwkKeyNotFound = &errors.ErrDescriptor{
	MessageFormat: "NwkKey not found for device",
	Type:          errors.NotFound,
	Code:          8,
}

// ErrAppKeyEnvelopeNotFound represents error ocurring when AppKey envelope was not found for device.
var ErrAppKeyEnvelopeNotFound = &errors.ErrDescriptor{
	MessageFormat: "AppKey envelope not found for device",
	Type:          errors.NotFound,
	Code:          9,
}

// ErrNwkKeyEnvelopeNotFound represents error ocurring when NwkKey envelope was not found for device.
var ErrNwkKeyEnvelopeNotFound = &errors.ErrDescriptor{
	MessageFormat: "NwkKey envelope not found for device",
	Type:          errors.NotFound,
	Code:          10,
}

// ErrMICCheckFailed represents error ocurring when MIC check failed.
var ErrMICCheckFailed = &errors.ErrDescriptor{
	MessageFormat: "MIC check failed",
	Type:          errors.Unknown,
	Code:          11,
}

// ErrUnsupportedLoRaWANVersion represents error ocurring when unsupported LoRaWAN MAC version is specified.
var ErrUnsupportedLoRaWANVersion = &errors.ErrDescriptor{
	MessageFormat: "Unsupported LoRaWAN MAC version: {lorawan_version}",
	Type:          errors.NotImplemented,
	Code:          12,
}

// ErrMissingDevAddr represents error ocurring when DevAddr is missing.
var ErrMissingDevAddr = &errors.ErrDescriptor{
	MessageFormat: "DevAddr is missing",
	Type:          errors.InvalidArgument,
	Code:          13,
}

// ErrMissingJoinEUI represents error ocurring when JoinEUI is missing.
var ErrMissingJoinEUI = &errors.ErrDescriptor{
	MessageFormat: "JoinEUI is missing",
	Type:          errors.InvalidArgument,
	Code:          14,
}

// ErrMissingDevEUI represents error ocurring when DevEUI is missing.
var ErrMissingDevEUI = &errors.ErrDescriptor{
	MessageFormat: "DevEUI is missing",
	Type:          errors.InvalidArgument,
	Code:          15,
}

// ErrMissingJoinRequest represents error ocurring when join request is missing.
var ErrMissingJoinRequest = &errors.ErrDescriptor{
	MessageFormat: "Join request is missing",
	Type:          errors.InvalidArgument,
	Code:          16,
}

// ErrUnmarshalFailed represents error ocurring when payload unmarshaling fails.
var ErrUnmarshalFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to unmarshal payload",
	Type:          errors.InvalidArgument,
	Code:          17,
}

// ErrForwardJoinRequest represents error ocurring when forwarding requests to other join servers is not implemented yet.
var ErrForwardJoinRequest = &errors.ErrDescriptor{
	MessageFormat: "Forwarding requests to other join servers is not implemented yet",
	Type:          errors.NotImplemented,
	Code:          18,
}

// ErrEncodeMHDRFailed represents error ocurring when encoding of join accept MHDR fails.
var ErrEncodeMHDRFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to encode join accept MHDR",
	Type:          errors.Unknown,
	Code:          19,
}

// ErrEncodePayloadFailed represents error ocurring when encodin of join accept payload fails.
var ErrEncodePayloadFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to encode join accept payload",
	Type:          errors.Unknown,
	Code:          20,
}

// ErrComputeJoinAcceptMIC represents error ocurring when computation of join accept MIC fails.
var ErrComputeJoinAcceptMIC = &errors.ErrDescriptor{
	MessageFormat: "Failed to compute join accept MIC",
	Type:          errors.Unknown,
	Code:          21,
}

// ErrEncryptPayloadFailed represents error ocurring when encryption of join accept fails.
var ErrEncryptPayloadFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to encrypt join accept",
	Type:          errors.Unknown,
	Code:          22,
}

// ErrDevNonceTooSmall represents error ocurring when DevNonce is too small.
var ErrDevNonceTooSmall = &errors.ErrDescriptor{
	MessageFormat: "DevNonce is too small",
	Type:          errors.InvalidArgument,
	Code:          23,
}

// ErrDevNonceReused represents error ocurring when DevNonce has already been used.
var ErrDevNonceReused = &errors.ErrDescriptor{
	MessageFormat: "DevNonce has already been used",
	Type:          errors.InvalidArgument,
	Code:          24,
}

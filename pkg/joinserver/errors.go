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

var ErrDeviceNotFound = &errors.ErrDescriptor{
	MessageFormat: "Device not found",
	Type:          errors.NotFound,
	Code:          1,
}

var ErrTooManyDevices = &errors.ErrDescriptor{
	MessageFormat: "Too many devices are associated with identifiers specified",
	Type:          errors.Conflict,
	Code:          2,
}

var ErrWrongPayloadType = &errors.ErrDescriptor{
	MessageFormat: "Wrong payload type: expected {expected_value}, got {got_value}",
	Type:          errors.InvalidArgument,
	Code:          3,
}

var ErrMissingPayload = &errors.ErrDescriptor{
	MessageFormat: "Join request payload is missing",
	Type:          errors.InvalidArgument,
	Code:          4,
}

var ErrWrongMICLength = &errors.ErrDescriptor{
	MessageFormat: "Wrong MIC length: expected 4, got {got_value}",
	Type:          errors.InvalidArgument,
	Code:          5,
}

var ErrInvalidMIC = &errors.ErrDescriptor{
	MessageFormat: "MIC mismatch",
	Type:          errors.InvalidArgument,
	Code:          6,
}

var ErrAppKeyNotFound = &errors.ErrDescriptor{
	MessageFormat: "AppKey not found for device",
	Type:          errors.NotFound,
	Code:          7,
}

var ErrNwkKeyNotFound = &errors.ErrDescriptor{
	MessageFormat: "NwkKey not found for device",
	Type:          errors.NotFound,
	Code:          8,
}

var ErrAppKeyEnvelopeNotFound = &errors.ErrDescriptor{
	MessageFormat: "AppKey envelope not found for device",
	Type:          errors.NotFound,
	Code:          9,
}

var ErrNwkKeyEnvelopeNotFound = &errors.ErrDescriptor{
	MessageFormat: "NwkKey envelope not found for device",
	Type:          errors.NotFound,
	Code:          10,
}

var ErrMICCheckFailed = &errors.ErrDescriptor{
	MessageFormat: "MIC check failed",
	Type:          errors.Unknown,
	Code:          11,
}

var ErrUnsupportedLoRaWANVersion = &errors.ErrDescriptor{
	MessageFormat: "Unsupported LoRaWAN MAC version: {lorawan_version}",
	Type:          errors.NotImplemented,
	Code:          12,
}

var ErrMissingDevAddr = &errors.ErrDescriptor{
	MessageFormat: "DevAddr is missing",
	Type:          errors.InvalidArgument,
	Code:          13,
}

var ErrMissingJoinEUI = &errors.ErrDescriptor{
	MessageFormat: "JoinEUI is missing",
	Type:          errors.InvalidArgument,
	Code:          14,
}

var ErrMissingDevEUI = &errors.ErrDescriptor{
	MessageFormat: "DevEUI is missing",
	Type:          errors.InvalidArgument,
	Code:          15,
}

var ErrMissingJoinRequest = &errors.ErrDescriptor{
	MessageFormat: "Join request is missing",
	Type:          errors.InvalidArgument,
	Code:          16,
}

var ErrUnmarshalFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to unmarshal payload specified",
	Type:          errors.InvalidArgument,
	Code:          17,
}

var ErrForwardJoinRequest = &errors.ErrDescriptor{
	MessageFormat: "Forwarding requests to other join servers is not implemented yet",
	Type:          errors.NotImplemented,
	Code:          18,
}

var ErrEncodeMHDRFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to encode join accept MHDR",
	Type:          errors.Unknown,
	Code:          19,
}

var ErrEncodePayloadFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to encode join accept payload",
	Type:          errors.Unknown,
	Code:          20,
}

var ErrComputeJoinAcceptMIC = &errors.ErrDescriptor{
	MessageFormat: "Failed to compute join accept MIC",
	Type:          errors.Unknown,
	Code:          21,
}

var ErrEncryptPayloadFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to encrypt join accept",
	Type:          errors.Unknown,
	Code:          22,
}

var ErrDevNonceTooSmall = &errors.ErrDescriptor{
	MessageFormat: "DevNonce is too small",
	Type:          errors.InvalidArgument,
	Code:          23,
}

var ErrDevNonceReused = &errors.ErrDescriptor{
	MessageFormat: "DevNonce already used",
	Type:          errors.InvalidArgument,
	Code:          24,
}

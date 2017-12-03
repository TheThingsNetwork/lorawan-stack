// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package joinserver

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrMICComputeFailed.Register()
	ErrUnsupportedLoRaWANMajorVersion.Register()
	ErrWrongPayloadType.Register()
	ErrMissingPayload.Register()
	ErrMICMismatch.Register()
	ErrAppKeyNotFound.Register()
	ErrNwkKeyNotFound.Register()
	ErrAppKeyEnvelopeNotFound.Register()
	ErrNwkKeyEnvelopeNotFound.Register()
	ErrMICCheckFailed.Register()
	ErrUnsupportedLoRaWANMACVersion.Register()
	ErrMissingDevAddr.Register()
	ErrMissingJoinEUI.Register()
	ErrMissingDevEUI.Register()
	ErrMissingJoinRequest.Register()
	ErrUnmarshalFailed.Register()
	ErrForwardJoinRequest.Register()
	ErrEncryptPayloadFailed.Register()
	ErrDevNonceTooSmall.Register()
	ErrDevNonceReused.Register()
	ErrCorruptRegistry.Register()
	ErrMACVersionMismatch.Register()
}

// ErrMICComputeFailed represents error occurring when MIC computation fails
var ErrMICComputeFailed = &errors.ErrDescriptor{
	MessageFormat: "Failed to compute MIC",
	Type:          errors.InvalidArgument,
	Code:          1,
}

// ErrUnsupportedLoRaWANMajorVersion represents error ocurring when unsupported LoRaWAN MAC version is specified.
var ErrUnsupportedLoRaWANMajorVersion = &errors.ErrDescriptor{
	MessageFormat: "Unsupported LoRaWAN major version: {major}",
	Type:          errors.NotImplemented,
	Code:          2,
}

// ErrWrongPayloadType represents error ocurring when wrong payload type is received.
var ErrWrongPayloadType = &errors.ErrDescriptor{
	MessageFormat:  "Wrong payload type: {type}",
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"type"},
	Code:           3,
}

// ErrMissingPayload represents error ocurring when join request payload is missing.
var ErrMissingPayload = &errors.ErrDescriptor{
	MessageFormat: "Join request payload is missing",
	Type:          errors.InvalidArgument,
	Code:          4,
}

// ErrMICMismatch represents error ocurring when MIC mismatch.
var ErrMICMismatch = &errors.ErrDescriptor{
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

// ErrUnsupportedLoRaWANMACVersion represents error ocurring when unsupported LoRaWAN MAC version is specified.
var ErrUnsupportedLoRaWANMACVersion = &errors.ErrDescriptor{
	MessageFormat: "Unsupported LoRaWAN MAC version: {version}",
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

// ErrCorruptRegistry represents error ocurring when join server registry is corrupted.
var ErrCorruptRegistry = &errors.ErrDescriptor{
	MessageFormat: "Registry is corrupt",
	Type:          errors.Internal,
	Code:          25,
}

// ErrMACVersionMismatch represents error ocurring when selected MAC version does not match version stored in join server registry.
var ErrMACVersionMismatch = &errors.ErrDescriptor{
	MessageFormat: "Device MAC version mismatch, in registry: {registered}, selected: {selected}",
	Type:          errors.Internal,
	Code:          26,
}

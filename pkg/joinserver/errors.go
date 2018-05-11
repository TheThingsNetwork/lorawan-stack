// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package joinserver

import "go.thethings.network/lorawan-stack/pkg/errors"

var (
	// ErrMICComputeFailed represents error occurring when MIC computation fails.
	ErrMICComputeFailed = &errors.ErrDescriptor{
		MessageFormat: "Failed to compute MIC",
		Type:          errors.InvalidArgument,
		Code:          1,
	}

	// ErrUnsupportedLoRaWANMajorVersion represents error ocurring when unsupported LoRaWAN MAC version is specified.
	ErrUnsupportedLoRaWANMajorVersion = &errors.ErrDescriptor{
		MessageFormat: "Unsupported LoRaWAN major version: {major}",
		Type:          errors.InvalidArgument,
		Code:          2,
	}

	// ErrWrongPayloadType represents error ocurring when wrong payload type is received.
	ErrWrongPayloadType = &errors.ErrDescriptor{
		MessageFormat:  "Wrong payload type: {type}",
		Type:           errors.InvalidArgument,
		SafeAttributes: []string{"type"},
		Code:           3,
	}

	// ErrMissingPayload represents error ocurring when join request payload is missing.
	ErrMissingPayload = &errors.ErrDescriptor{
		MessageFormat: "Join request payload is missing",
		Type:          errors.InvalidArgument,
		Code:          4,
	}

	// ErrMICMismatch represents error ocurring when MIC mismatch.
	ErrMICMismatch = &errors.ErrDescriptor{
		MessageFormat: "MIC mismatch",
		Type:          errors.InvalidArgument,
		Code:          5,
	}

	// ErrAppKeyNotFound represents error ocurring when AppKey was not found for device.
	ErrAppKeyNotFound = &errors.ErrDescriptor{
		MessageFormat: "AppKey not found for device",
		Type:          errors.NotFound,
		Code:          6,
	}

	// ErrNwkKeyNotFound represents error ocurring when NwkKey was not found for device.
	ErrNwkKeyNotFound = &errors.ErrDescriptor{
		MessageFormat: "NwkKey not found for device",
		Type:          errors.NotFound,
		Code:          7,
	}

	// ErrAppKeyEnvelopeNotFound represents error ocurring when AppKey envelope was not found for device.
	ErrAppKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "AppKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          8,
	}

	// ErrNwkKeyEnvelopeNotFound represents error ocurring when NwkKey envelope was not found for device.
	ErrNwkKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "NwkKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          9,
	}

	// ErrMICCheckFailed represents error ocurring when MIC check failed.
	ErrMICCheckFailed = &errors.ErrDescriptor{
		MessageFormat: "MIC check failed",
		Type:          errors.Unknown,
		Code:          10,
	}

	// ErrUnsupportedLoRaWANMACVersion represents error ocurring when unsupported LoRaWAN MAC version is specified.
	ErrUnsupportedLoRaWANMACVersion = &errors.ErrDescriptor{
		MessageFormat: "Unsupported LoRaWAN MAC version: {version}",
		Type:          errors.InvalidArgument,
		Code:          11,
	}

	// ErrMissingDevAddr represents error ocurring when DevAddr is missing.
	ErrMissingDevAddr = &errors.ErrDescriptor{
		MessageFormat: "DevAddr is missing",
		Type:          errors.InvalidArgument,
		Code:          12,
	}

	// ErrMissingJoinEUI represents error ocurring when JoinEUI is missing.
	ErrMissingJoinEUI = &errors.ErrDescriptor{
		MessageFormat: "JoinEUI is missing",
		Type:          errors.InvalidArgument,
		Code:          13,
	}

	// ErrMissingDevEUI represents error ocurring when DevEUI is missing.
	ErrMissingDevEUI = &errors.ErrDescriptor{
		MessageFormat: "DevEUI is missing",
		Type:          errors.InvalidArgument,
		Code:          14,
	}

	// ErrMissingJoinRequest represents error ocurring when join request is missing.
	ErrMissingJoinRequest = &errors.ErrDescriptor{
		MessageFormat: "Join request is missing",
		Type:          errors.InvalidArgument,
		Code:          15,
	}

	// ErrUnmarshalFailed represents error ocurring when payload unmarshaling fails.
	ErrUnmarshalFailed = &errors.ErrDescriptor{
		MessageFormat: "Failed to unmarshal payload",
		Type:          errors.InvalidArgument,
		Code:          16,
	}

	// ErrForwardJoinRequest represents error ocurring when forwarding join request.
	ErrForwardJoinRequest = &errors.ErrDescriptor{
		MessageFormat: "Failed to forward join request",
		Type:          errors.Unknown,
		Code:          17,
	}

	// ErrComputeJoinAcceptMIC represents error ocurring when computation of join accept MIC fails.
	ErrComputeJoinAcceptMIC = &errors.ErrDescriptor{
		MessageFormat: "Failed to compute join accept MIC",
		Type:          errors.Unknown,
		Code:          18,
	}

	// ErrEncryptPayloadFailed represents error ocurring when encryption of join accept fails.
	ErrEncryptPayloadFailed = &errors.ErrDescriptor{
		MessageFormat: "Failed to encrypt join accept",
		Type:          errors.Unknown,
		Code:          19,
	}

	// ErrDevNonceTooSmall represents error ocurring when DevNonce is too small.
	ErrDevNonceTooSmall = &errors.ErrDescriptor{
		MessageFormat: "DevNonce is too small",
		Type:          errors.InvalidArgument,
		Code:          20,
	}

	// ErrDevNonceReused represents error ocurring when DevNonce has already been used.
	ErrDevNonceReused = &errors.ErrDescriptor{
		MessageFormat: "DevNonce has already been used",
		Type:          errors.InvalidArgument,
		Code:          21,
	}

	// ErrCorruptRegistry represents error ocurring when join server registry is corrupted.
	ErrCorruptRegistry = &errors.ErrDescriptor{
		MessageFormat: "Registry is corrupt",
		Type:          errors.Internal,
		Code:          22,
	}

	// ErrMACVersionMismatch represents error ocurring when selected MAC version does not match version stored in join server registry.
	ErrMACVersionMismatch = &errors.ErrDescriptor{
		MessageFormat: "Device MAC version mismatch, in registry: {registered}, selected: {selected}",
		Type:          errors.Internal,
		Code:          23,
	}

	// ErrDevNonceTooHigh represents error ocurring when DevNonce is too high.
	ErrDevNonceTooHigh = &errors.ErrDescriptor{
		MessageFormat: "DevNonce is too high",
		Type:          errors.InvalidArgument,
		Code:          24,
	}

	// ErrAddressMismatch represents error ocurring when the address of a component does not match the one associated with the device.
	ErrAddressMismatch = &errors.ErrDescriptor{
		MessageFormat: "{component} address mismatch",
		Type:          errors.PermissionDenied,
		Code:          25,
	}

	// ErrNoSession represents error ocurring when there is no session associated with the device.
	ErrNoSession = &errors.ErrDescriptor{
		MessageFormat: "Device has no session associated with it",
		Type:          errors.NotFound,
		Code:          26,
	}

	// ErrSessionKeyIDMismatch represents error ocurring when the specified Session Key ID does not match any of the ones associated with the device.
	ErrSessionKeyIDMismatch = &errors.ErrDescriptor{
		MessageFormat: "Session key ID mismatch",
		Type:          errors.InvalidArgument,
		Code:          27,
	}

	// ErrAppSKeyEnvelopeNotFound represents error ocurring when AppSKey envelope was not found for device.
	ErrAppSKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "AppSKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          28,
	}

	// ErrNwkSEncKeyEnvelopeNotFound represents error ocurring when NwkSEncKey envelope was not found for device.
	ErrNwkSEncKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "NwkSEncKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          29,
	}

	// ErrFNwkSIntKeyEnvelopeNotFound represents error ocurring when FNwkSIntKey envelope was not found for device.
	ErrFNwkSIntKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "FNwkSIntKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          30,
	}

	// ErrSNwkSIntKeyEnvelopeNotFound represents error ocurring when SNwkSIntKey envelope was not found for device.
	ErrSNwkSIntKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "SNwkSIntKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          31,
	}

	// ErrMissingSessionKeyID represents error ocurring when SessionKeyID is missing.
	ErrMissingSessionKeyID = &errors.ErrDescriptor{
		MessageFormat: "SessionKeyID is missing",
		Type:          errors.InvalidArgument,
		Code:          32,
	}

	ErrUnknownAppEUI = &errors.ErrDescriptor{
		MessageFormat: "AppEUI specified is not known",
		Type:          errors.Unknown,
		Code:          33,
	}
)

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
	ErrAddressMismatch.Register()
	ErrNoSession.Register()
	ErrAppSKeyEnvelopeNotFound.Register()
	ErrNwkSEncKeyEnvelopeNotFound.Register()
	ErrFNwkSIntKeyEnvelopeNotFound.Register()
	ErrSNwkSIntKeyEnvelopeNotFound.Register()
	ErrMissingSessionKeyID.Register()
	ErrSessionKeyIDMismatch.Register()
	ErrUnknownAppEUI.Register()
}

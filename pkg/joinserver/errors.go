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

	// ErrWrongPayloadType represents error occurring when wrong payload type is received.
	ErrWrongPayloadType = &errors.ErrDescriptor{
		MessageFormat:  "Wrong payload type: {type}",
		Type:           errors.InvalidArgument,
		SafeAttributes: []string{"type"},
		Code:           2,
	}

	// ErrMICMismatch represents error occurring when MIC mismatch.
	ErrMICMismatch = &errors.ErrDescriptor{
		MessageFormat: "MIC mismatch",
		Type:          errors.InvalidArgument,
		Code:          3,
	}

	// ErrAppKeyNotFound represents error occurring when AppKey was not found for device.
	ErrAppKeyNotFound = &errors.ErrDescriptor{
		MessageFormat: "AppKey not found for device",
		Type:          errors.NotFound,
		Code:          4,
	}

	// ErrNwkKeyNotFound represents error occurring when NwkKey was not found for device.
	ErrNwkKeyNotFound = &errors.ErrDescriptor{
		MessageFormat: "NwkKey not found for device",
		Type:          errors.NotFound,
		Code:          5,
	}

	// ErrAppKeyEnvelopeNotFound represents error occurring when AppKey envelope was not found for device.
	ErrAppKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "AppKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          6,
	}

	// ErrNwkKeyEnvelopeNotFound represents error occurring when NwkKey envelope was not found for device.
	ErrNwkKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "NwkKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          7,
	}

	// ErrMICCheckFailed represents error occurring when MIC check failed.
	ErrMICCheckFailed = &errors.ErrDescriptor{
		MessageFormat: "MIC check failed",
		Type:          errors.Unknown,
		Code:          8,
	}

	// ErrMissingJoinRequest represents error occurring when join request is missing.
	ErrMissingJoinRequest = &errors.ErrDescriptor{
		MessageFormat: "Join request is missing",
		Type:          errors.InvalidArgument,
		Code:          9,
	}

	// ErrForwardJoinRequest represents error occurring when forwarding join request.
	ErrForwardJoinRequest = &errors.ErrDescriptor{
		MessageFormat: "Failed to forward join request",
		Type:          errors.Unknown,
		Code:          10,
	}

	// ErrEncryptPayloadFailed represents error occurring when encryption of join accept fails.
	ErrEncryptPayloadFailed = &errors.ErrDescriptor{
		MessageFormat: "Failed to encrypt join accept",
		Type:          errors.Unknown,
		Code:          11,
	}

	// ErrDevNonceTooSmall represents error occurring when DevNonce is too small.
	ErrDevNonceTooSmall = &errors.ErrDescriptor{
		MessageFormat: "DevNonce is too small",
		Type:          errors.InvalidArgument,
		Code:          12,
	}

	// ErrDevNonceReused represents error occurring when DevNonce has already been used.
	ErrDevNonceReused = &errors.ErrDescriptor{
		MessageFormat: "DevNonce has already been used",
		Type:          errors.InvalidArgument,
		Code:          13,
	}

	// ErrMACVersionMismatch represents error occurring when selected MAC version does not match version stored in Join Server registry.
	ErrMACVersionMismatch = &errors.ErrDescriptor{
		MessageFormat: "Device MAC version mismatch, in registry: {registered}, selected: {selected}",
		Type:          errors.Internal,
		Code:          14,
	}

	// ErrDevNonceTooHigh represents error occurring when DevNonce is too high.
	ErrDevNonceTooHigh = &errors.ErrDescriptor{
		MessageFormat: "DevNonce is too high",
		Type:          errors.InvalidArgument,
		Code:          15,
	}

	// ErrAddressMismatch represents error occurring when the address of a component does not match the one associated with the device.
	ErrAddressMismatch = &errors.ErrDescriptor{
		MessageFormat: "{component} address mismatch",
		Type:          errors.PermissionDenied,
		Code:          16,
	}

	// ErrNoSession represents error occurring when there is no session associated with the device.
	ErrNoSession = &errors.ErrDescriptor{
		MessageFormat: "Device has no session associated with it",
		Type:          errors.NotFound,
		Code:          17,
	}

	// ErrSessionKeyIDMismatch represents error occurring when the specified Session Key ID does not match any of the ones associated with the device.
	ErrSessionKeyIDMismatch = &errors.ErrDescriptor{
		MessageFormat: "Session key ID mismatch",
		Type:          errors.InvalidArgument,
		Code:          18,
	}

	// ErrAppSKeyEnvelopeNotFound represents error occurring when AppSKey envelope was not found for device.
	ErrAppSKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "AppSKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          19,
	}

	// ErrNwkSEncKeyEnvelopeNotFound represents error occurring when NwkSEncKey envelope was not found for device.
	ErrNwkSEncKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "NwkSEncKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          20,
	}

	// ErrFNwkSIntKeyEnvelopeNotFound represents error occurring when FNwkSIntKey envelope was not found for device.
	ErrFNwkSIntKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "FNwkSIntKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          21,
	}

	// ErrSNwkSIntKeyEnvelopeNotFound represents error occurring when SNwkSIntKey envelope was not found for device.
	ErrSNwkSIntKeyEnvelopeNotFound = &errors.ErrDescriptor{
		MessageFormat: "SNwkSIntKey envelope not found for device",
		Type:          errors.NotFound,
		Code:          22,
	}

	// ErrMissingSessionKeyID represents error occurring when SessionKeyID is missing.
	ErrMissingSessionKeyID = &errors.ErrDescriptor{
		MessageFormat: "SessionKeyID is missing",
		Type:          errors.InvalidArgument,
		Code:          23,
	}

	// ErrUnknownAppEUI represents the error occurring when the AppEUI specified is unknown.
	ErrUnknownAppEUI = &errors.ErrDescriptor{
		MessageFormat: "AppEUI specified is not known",
		Type:          errors.Unknown,
		Code:          24,
	}
)

func init() {
	ErrMICComputeFailed.Register()
	ErrWrongPayloadType.Register()
	ErrMICMismatch.Register()
	ErrAppKeyNotFound.Register()
	ErrNwkKeyNotFound.Register()
	ErrAppKeyEnvelopeNotFound.Register()
	ErrNwkKeyEnvelopeNotFound.Register()
	ErrMICCheckFailed.Register()
	ErrMissingJoinRequest.Register()
	ErrForwardJoinRequest.Register()
	ErrEncryptPayloadFailed.Register()
	ErrDevNonceTooSmall.Register()
	ErrDevNonceReused.Register()
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

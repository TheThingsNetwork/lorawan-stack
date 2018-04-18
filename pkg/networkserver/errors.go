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

package networkserver

import "github.com/TheThingsNetwork/ttn/pkg/errors"

var (
	// ErrUnsupportedLoRaWANMajorVersion represents error ocurring when unsupported LoRaWAN MAC version is specified.
	ErrUnsupportedLoRaWANMajorVersion = &errors.ErrDescriptor{
		MessageFormat:  "Unsupported LoRaWAN major version: `{major}`",
		Type:           errors.InvalidArgument,
		Code:           1,
		SafeAttributes: []string{"major"},
	}

	// ErrMissingPayload represents error ocurring when message payload is missing.
	ErrMissingPayload = &errors.ErrDescriptor{
		MessageFormat: "Message payload is missing",
		Type:          errors.InvalidArgument,
		Code:          2,
	}

	// ErrUnmarshalFailed represents error ocurring when payload unmarshaling fails.
	ErrUnmarshalFailed = &errors.ErrDescriptor{
		MessageFormat: "Failed to unmarshal payload",
		Type:          errors.InvalidArgument,
		Code:          3,
	}

	// ErrFCntTooLow represents error ocurring when FCnt is too low.
	ErrFCntTooLow = &errors.ErrDescriptor{
		MessageFormat: "FCnt is too low",
		Type:          errors.InvalidArgument,
		Code:          4,
	}

	// ErrFCntTooHigh represents error ocurring when FCnt is too high.
	ErrFCntTooHigh = &errors.ErrDescriptor{
		MessageFormat: "FCnt is too high",
		Type:          errors.InvalidArgument,
		Code:          5,
	}

	// ErrCorruptRegistry represents error ocurring when network server registry is corrupt.
	ErrCorruptRegistry = &errors.ErrDescriptor{
		MessageFormat: "Registry is corrupt",
		Type:          errors.Internal,
		Code:          6,
	}

	// ErrMICComputeFailed represents error ocurring when MIC computation fails.
	ErrMICComputeFailed = &errors.ErrDescriptor{
		MessageFormat: "Failed to compute MIC",
		Type:          errors.InvalidArgument,
		Code:          7,
	}

	// ErrDeviceNotFound represents error ocurring when device is not found.
	ErrDeviceNotFound = &errors.ErrDescriptor{
		MessageFormat: "Device not found",
		Type:          errors.NotFound,
		Code:          8,
	}

	// ErrMissingFNwkSIntKey represents error ocurring when FNwkSIntKey is missing.
	ErrMissingFNwkSIntKey = &errors.ErrDescriptor{
		MessageFormat: "FNwkSIntKey is unknown",
		Type:          errors.NotFound,
		Code:          9,
	}

	// ErrMissingSNwkSIntKey represents error ocurring when SNwkSIntKey is missing.
	ErrMissingSNwkSIntKey = &errors.ErrDescriptor{
		MessageFormat: "SNwkSIntKey is unknown",
		Type:          errors.NotFound,
		Code:          10,
	}

	// ErrNewSubscription represents error ocurring when a new subscription is opened.
	ErrNewSubscription = &errors.ErrDescriptor{
		MessageFormat: "Another subscription started",
		Type:          errors.Conflict,
		Code:          11,
	}

	// ErrInvalidConfiguration represents error ocurring when the configuration is invalid.
	ErrInvalidConfiguration = &errors.ErrDescriptor{
		MessageFormat: "Invalid configuration",
		Type:          errors.InvalidArgument,
		Code:          12,
	}
)

func init() {
	ErrUnsupportedLoRaWANMajorVersion.Register()
	ErrUnmarshalFailed.Register()
	ErrFCntTooLow.Register()
	ErrFCntTooHigh.Register()
	ErrCorruptRegistry.Register()
	ErrMICComputeFailed.Register()
	ErrDeviceNotFound.Register()
	ErrMissingFNwkSIntKey.Register()
	ErrMissingSNwkSIntKey.Register()
	ErrNewSubscription.Register()
	ErrInvalidConfiguration.Register()
}

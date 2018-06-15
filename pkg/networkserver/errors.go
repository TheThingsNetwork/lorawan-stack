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

import "go.thethings.network/lorawan-stack/pkg/errors"

var (
	// ErrDeviceNotFound represents error ocurring when device is not found.
	ErrDeviceNotFound = &errors.ErrDescriptor{
		MessageFormat: "Device not found",
		Type:          errors.NotFound,
		Code:          1,
	}

	// ErrMissingFNwkSIntKey represents error ocurring when FNwkSIntKey is missing.
	ErrMissingFNwkSIntKey = &errors.ErrDescriptor{
		MessageFormat: "FNwkSIntKey is unknown",
		Type:          errors.NotFound,
		Code:          2,
	}

	// ErrMissingSNwkSIntKey represents error ocurring when SNwkSIntKey is missing.
	ErrMissingSNwkSIntKey = &errors.ErrDescriptor{
		MessageFormat: "SNwkSIntKey is unknown",
		Type:          errors.NotFound,
		Code:          3,
	}

	// ErrMissingNwkSEncKey represents error ocurring when NwkSEncKey is missing.
	ErrMissingNwkSEncKey = &errors.ErrDescriptor{
		MessageFormat: "NwkSEncKey is unknown",
		Type:          errors.NotFound,
		Code:          4,
	}

	// ErrMissingApplicationID represents error ocurring when ApplicationID is missing.
	ErrMissingApplicationID = &errors.ErrDescriptor{
		MessageFormat: "Application ID is unknown",
		Type:          errors.NotFound,
		Code:          5,
	}

	// ErrMissingGatewayID represents error ocurring when GatewayID is missing.
	ErrMissingGatewayID = &errors.ErrDescriptor{
		MessageFormat: "Gateway ID is unknown",
		Type:          errors.NotFound,
		Code:          6,
	}

	// ErrNewSubscription represents error ocurring when a new subscription is opened.
	ErrNewSubscription = &errors.ErrDescriptor{
		MessageFormat: "Another subscription started",
		Type:          errors.Conflict,
		Code:          7,
	}

	// ErrInvalidConfiguration represents error ocurring when the configuration is invalid.
	ErrInvalidConfiguration = &errors.ErrDescriptor{
		MessageFormat: "Invalid configuration",
		Type:          errors.InvalidArgument,
		Code:          8,
	}

	// ErrUplinkNotFound represents error ocurring when there were no uplinks found.
	ErrUplinkNotFound = &errors.ErrDescriptor{
		MessageFormat: "Uplink not found",
		Type:          errors.NotFound,
		Code:          9,
	}

	// ErrGatewayServerNotFound represents error ocurring when there were no uplinks found.
	ErrGatewayServerNotFound = &errors.ErrDescriptor{
		MessageFormat: "Gateway server not found",
		Type:          errors.NotFound,
		Code:          10,
	}

	// ErrChannelIndexTooHigh represents error ocurring when the channel index is too high.
	ErrChannelIndexTooHigh = &errors.ErrDescriptor{
		MessageFormat: "Channel index is too high",
		Type:          errors.InvalidArgument,
		Code:          11,
	}

	// ErrDecryptionFailed represents error ocurring when the decryption fails.
	ErrDecryptionFailed = &errors.ErrDescriptor{
		MessageFormat: "Decryption failed",
		Type:          errors.InvalidArgument,
		Code:          12,
	}

	// ErrMACRequestNotFound represents error ocurring when the a response to a MAC response
	// is received, but a corresponding request is not found.
	ErrMACRequestNotFound = &errors.ErrDescriptor{
		MessageFormat: "MAC response received, but corresponding request not found",
		Type:          errors.InvalidArgument,
		Code:          13,
	}

	// ErrInvalidDataRate represents error ocurring when the data rate is invalid.
	ErrInvalidDataRate = &errors.ErrDescriptor{
		MessageFormat: "Invalid data rate",
		Type:          errors.InvalidArgument,
		Code:          14,
	}

	// ErrScheduleTooSoon represents error ocurring when a confirmed downlink is scheduled too soon.
	ErrScheduleTooSoon = &errors.ErrDescriptor{
		MessageFormat: "Confirmed downlink is scheduled too soon",
		Type:          errors.TemporarilyUnavailable,
		Code:          15,
	}
)

func init() {
	ErrDeviceNotFound.Register()
	ErrMissingFNwkSIntKey.Register()
	ErrMissingSNwkSIntKey.Register()
	ErrMissingNwkSEncKey.Register()
	ErrMissingApplicationID.Register()
	ErrMissingGatewayID.Register()
	ErrNewSubscription.Register()
	ErrInvalidConfiguration.Register()
	ErrUplinkNotFound.Register()
	ErrGatewayServerNotFound.Register()
	ErrChannelIndexTooHigh.Register()
	ErrDecryptionFailed.Register()
	ErrMACRequestNotFound.Register()
	ErrInvalidDataRate.Register()
	ErrScheduleTooSoon.Register()
}

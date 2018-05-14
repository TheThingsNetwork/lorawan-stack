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

	// ErrMissingApplicationID represents error ocurring when ApplicationID is missing.
	ErrMissingApplicationID = &errors.ErrDescriptor{
		MessageFormat: "ApplicationID is unknown",
		Type:          errors.NotFound,
		Code:          4,
	}

	// ErrNewSubscription represents error ocurring when a new subscription is opened.
	ErrNewSubscription = &errors.ErrDescriptor{
		MessageFormat: "Another subscription started",
		Type:          errors.Conflict,
		Code:          5,
	}

	// ErrInvalidConfiguration represents error ocurring when the configuration is invalid.
	ErrInvalidConfiguration = &errors.ErrDescriptor{
		MessageFormat: "Invalid configuration",
		Type:          errors.InvalidArgument,
		Code:          6,
	}
)

func init() {
	ErrDeviceNotFound.Register()
	ErrMissingFNwkSIntKey.Register()
	ErrMissingSNwkSIntKey.Register()
	ErrMissingApplicationID.Register()
	ErrNewSubscription.Register()
	ErrInvalidConfiguration.Register()
}

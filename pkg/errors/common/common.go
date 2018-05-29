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

package common

import "go.thethings.network/lorawan-stack/pkg/errors"

var (
	// ErrInvalidArgument is returned if the arguments passed to a function are invalid.
	ErrInvalidArgument = &errors.ErrDescriptor{
		MessageFormat: "Invalid arguments",
		Code:          1,
		Type:          errors.InvalidArgument,
	}
	// ErrCheckFailed is returned if the arguments didn't pass a specifically-defined
	// argument check.
	ErrCheckFailed = &errors.ErrDescriptor{
		MessageFormat: "Arguments check failed",
		Code:          2,
		Type:          errors.InvalidArgument,
	}
	// ErrUnmarshalPayloadFailed is returned when a payload couldn't be unmarshalled.
	ErrUnmarshalPayloadFailed = &errors.ErrDescriptor{
		MessageFormat: "Failed to unmarshal payload",
		Type:          errors.InvalidArgument,
		Code:          3,
	}
	// ErrMarshalPayloadFailed is returned when a payload couldn't be marshalled.
	ErrMarshalPayloadFailed = &errors.ErrDescriptor{
		MessageFormat: "Failed to marshal payload",
		Type:          errors.InvalidArgument,
		Code:          17,
	}
	// ErrCorruptRegistry represents error occurring when the registry of a component is corrupted.
	ErrCorruptRegistry = &errors.ErrDescriptor{
		MessageFormat: "Registry is corrupt",
		Type:          errors.Internal,
		Code:          4,
	}
	// ErrPermissionDenied is returned when a request is not allowed to access a protected resource.
	ErrPermissionDenied = &errors.ErrDescriptor{
		MessageFormat: "Permission denied to perform this operation",
		Type:          errors.PermissionDenied,
		Code:          5,
	}
	// ErrUnauthorized is returned when a request has not been authorized to access a protected resource.
	ErrUnauthorized = &errors.ErrDescriptor{
		MessageFormat: "Unauthorized to perform this operation",
		Type:          errors.Unauthorized,
		Code:          16,
	}
)

func init() {
	ErrInvalidArgument.Register()
	ErrCheckFailed.Register()
	ErrUnmarshalPayloadFailed.Register()
	ErrMarshalPayloadFailed.Register()
	ErrCorruptRegistry.Register()
	ErrUnauthorized.Register()
	ErrPermissionDenied.Register()
}

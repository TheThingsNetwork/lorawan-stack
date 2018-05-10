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

package applicationregistry

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

var (
	// ErrApplicationNotFound represents the ErrDescriptor of the error returned
	// when the application is not found.
	ErrApplicationNotFound = &errors.ErrDescriptor{
		MessageFormat: "Application not found",
		Type:          errors.NotFound,
		Code:          1,
	}

	// ErrTooManyApplications represents the ErrDescriptor of the error returned
	// when there are too many applications associated with the identifiers specified.
	ErrTooManyApplications = &errors.ErrDescriptor{
		MessageFormat: "Too many applications found",
		Type:          errors.Conflict,
		Code:          2,
	}

	// ErrCheckFailed represents the ErrDescriptor of the error returned
	// when the check failed.
	ErrCheckFailed = &errors.ErrDescriptor{
		MessageFormat: "Argument check failed",
		Type:          errors.InvalidArgument,
		Code:          3,
	}
)

func init() {
	ErrApplicationNotFound.Register()
	ErrTooManyApplications.Register()
	ErrCheckFailed.Register()
}

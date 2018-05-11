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

package rights

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

var (
	// ErrPermissionDenied is returned if the rights were insufficient to perform
	// this operation.
	ErrPermissionDenied = &errors.ErrDescriptor{
		MessageFormat: "Permission denied to perform this operation",
		Type:          errors.PermissionDenied,
		Code:          1,
	}

	// ErrInvalidApplicationID is returned if an invalid application ID was passed to an
	// operation that requires it.
	ErrInvalidApplicationID = &errors.ErrDescriptor{
		MessageFormat: "Invalid application ID given",
		Type:          errors.InvalidArgument,
		Code:          2,
	}
)

func init() {
	ErrPermissionDenied.Register()
	ErrInvalidApplicationID.Register()
}

// RequireApplication checks, based on the context, that the clients has the
// specified rights for this application.
func RequireApplication(ctx context.Context, appIdentifiers ApplicationIDGetter, rights ...ttnpb.Right) error {
	// TODO: Accept administrator authorization even if not tied to the application
	// https://github.com/TheThingsIndustries/ttn/issues/731
	if err := validate.ID(appIdentifiers.GetApplicationID()); err != nil {
		return ErrInvalidApplicationID.NewWithCause(nil, err)
	}

	if ad := FromContext(ctx); !ttnpb.IncludesRights(ad, rights...) {
		return ErrPermissionDenied.New(nil)
	}

	return nil
}

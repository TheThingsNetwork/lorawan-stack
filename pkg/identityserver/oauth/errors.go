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

package oauth

import "go.thethings.network/lorawan-stack/pkg/errors"

func init() {
	ErrInvalidPassword.Register()
	ErrNotAuthenticated.Register()
	ErrInternal.Register()
}

// ErrInvalidPassword is returned by the login handler when providing a wrong
// password to login into an user's account.
var ErrInvalidPassword = &errors.ErrDescriptor{
	MessageFormat: "Invalid password",
	Code:          1,
	Type:          errors.InvalidArgument,
}

// ErrNotAuthenticated is returned when the user is not logged in but needs
// to be in order to perform a specific action.
var ErrNotAuthenticated = &errors.ErrDescriptor{
	MessageFormat: "Not authenticated",
	Code:          2,
	Type:          errors.Unauthorized,
}

// ErrInternal is returned by a handler when an underlying error happened that
// is not meant to be shown to the end-user.
var ErrInternal = &errors.ErrDescriptor{
	MessageFormat: "Internal error",
	Code:          3,
	Type:          errors.Internal,
}

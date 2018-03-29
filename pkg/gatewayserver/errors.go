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

package gatewayserver

import "github.com/TheThingsNetwork/ttn/pkg/errors"

var (
	// ErrNoNetworkServerFound is returned if no network server was found for a passed DevAddr.
	ErrNoNetworkServerFound = &errors.ErrDescriptor{
		MessageFormat: "No network server found for this message",
		Code:          1,
		Type:          errors.NotFound,
	}
	// ErrNoIdentityServerFound is returned if no identity server was found.
	ErrNoIdentityServerFound = &errors.ErrDescriptor{
		MessageFormat: "No identity server found",
		Code:          2,
		Type:          errors.NotFound,
	}
	// ErrUnauthorized is returned if there are no credentials passed.
	ErrUnauthorized = &errors.ErrDescriptor{
		MessageFormat: "No credentials passed",
		Code:          3,
		Type:          errors.Unauthorized,
	}
	// ErrPermissionDenied is returned if the credentials passed do not have enough rights to exchange gateway traffic.
	ErrPermissionDenied = &errors.ErrDescriptor{
		MessageFormat: "Not have enough rights to exchange gateway traffic",
		Code:          4,
		Type:          errors.PermissionDenied,
	}
)

func init() {
	ErrNoNetworkServerFound.Register()
	ErrNoIdentityServerFound.Register()
	ErrPermissionDenied.Register()
	ErrUnauthorized.Register()
}

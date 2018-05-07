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

package deviceregistry

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ErrDeviceNotFound represents the ErrDescriptor of the error returned
// when the device is not found.
var ErrDeviceNotFound = &errors.ErrDescriptor{
	MessageFormat: "Device not found",
	Type:          errors.NotFound,
	Code:          1,
}

// ErrTooManyDevices represents the ErrDescriptor of the error returned
// when there are too many devices associated with the identifiers specified.
var ErrTooManyDevices = &errors.ErrDescriptor{
	MessageFormat: "Too many devices found",
	Type:          errors.Conflict,
	Code:          2,
}

// ErrCheckFailed represents the ErrDescriptor of the error returned
// when the check failed.
var ErrCheckFailed = &errors.ErrDescriptor{
	MessageFormat: "Argument check failed",
	Type:          errors.InvalidArgument,
	Code:          3,
}

// ErrPermissionDenied is returned if the rights were insufficient to perform
// this operation.
var ErrPermissionDenied = &errors.ErrDescriptor{
	MessageFormat: "Permission denied to perform this operation",
	Type:          errors.PermissionDenied,
	Code:          4,
}

// ErrNoApplicationID is returned if no application ID was passed to an
// operation that requires it.
var ErrNoApplicationID = &errors.ErrDescriptor{
	MessageFormat: "No application ID given",
	Type:          errors.InvalidArgument,
	Code:          5,
}

var componentsDiminutives = map[ttnpb.PeerInfo_Role]string{
	ttnpb.PeerInfo_APPLICATION_SERVER: "As",
	ttnpb.PeerInfo_NETWORK_SERVER:     "Ns",
	ttnpb.PeerInfo_JOIN_SERVER:        "Js",
}

func init() {
	ErrDeviceNotFound.Register()
	ErrTooManyDevices.Register()
	ErrCheckFailed.Register()
	ErrPermissionDenied.Register()
	ErrNoApplicationID.Register()
}

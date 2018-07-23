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

// Package unique provides functionality for working with unique identifiers of entities within a context.
package unique

import (
	"context"
	"fmt"
	"strings"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// ID returns the unique identifier of the specified identifiers.
func ID(ctx context.Context, id ttnpb.Identifiers) string {
	if id == nil {
		return ""
	}
	switch val := id.(type) {
	case ttnpb.ApplicationIdentifiers:
		return val.ApplicationID
	case *ttnpb.ApplicationIdentifiers:
		return val.GetApplicationID()
	case ttnpb.ClientIdentifiers:
		return val.ClientID
	case *ttnpb.ClientIdentifiers:
		return val.GetClientID()
	case ttnpb.EndDeviceIdentifiers:
		return fmt.Sprintf("%v:%v", val.ApplicationID, val.DeviceID)
	case *ttnpb.EndDeviceIdentifiers:
		if val == nil {
			return ""
		}
		return fmt.Sprintf("%v:%v", val.ApplicationID, val.DeviceID)
	case ttnpb.GatewayIdentifiers:
		return val.GatewayID
	case *ttnpb.GatewayIdentifiers:
		return val.GetGatewayID()
	case ttnpb.OrganizationIdentifiers:
		return val.OrganizationID
	case *ttnpb.OrganizationIdentifiers:
		return val.GetOrganizationID()
	case ttnpb.UserIdentifiers:
		return val.UserID
	case *ttnpb.UserIdentifiers:
		return val.GetUserID()
	default:
		panic(fmt.Errorf("Could not determine unique ID: %T is not a valid ttnpb.Identifiers", id))
	}
}

var errFormat = errors.DefineInvalidArgument("format", "invalid format in value `{value}`")

// ToApplicationID returns the application identifier of the specified unique ID.
func ToApplicationID(uid string) (ttnpb.ApplicationIdentifiers, error) {
	return ttnpb.ApplicationIdentifiers{ApplicationID: uid}, nil
}

// ToClientID returns the client identifier of the specified unique ID.
func ToClientID(uid string) (ttnpb.ClientIdentifiers, error) {
	return ttnpb.ClientIdentifiers{ClientID: uid}, nil
}

// ToDeviceID returns the end device identifier of the specified unique ID.
func ToDeviceID(uid string) (id ttnpb.EndDeviceIdentifiers, err error) {
	if parts := strings.SplitN(uid, ":", 2); len(parts) == 2 {
		id.ApplicationID = parts[0]
		id.DeviceID = parts[1]
	} else {
		err = errFormat.WithAttributes("value", uid)
	}
	return
}

// ToGatewayID returns the gateway identifier of the specified unique ID.
func ToGatewayID(uid string) (ttnpb.GatewayIdentifiers, error) {
	return ttnpb.GatewayIdentifiers{GatewayID: uid}, nil
}

// ToOrganizationID returns the organization identifier of the specified unique ID.
func ToOrganizationID(uid string) (ttnpb.OrganizationIdentifiers, error) {
	return ttnpb.OrganizationIdentifiers{OrganizationID: uid}, nil
}

// ToUserID returns the user identifier of the specified unique ID.
func ToUserID(uid string) (ttnpb.UserIdentifiers, error) {
	return ttnpb.UserIdentifiers{UserID: uid}, nil
}

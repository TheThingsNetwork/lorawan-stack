// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const separator = "."

var errUniqueIdentifier = errors.DefineInvalidArgument("unique_identifier", "invalid unique identifier `{uid}`")
var errFormat = errors.DefineInvalidArgument("format", "invalid format in value `{value}`")

// ID returns the unique identifier of the specified identifiers.
// This function panics if id is nil, if it's zero, or if it's is not a
// built-in identifiers type: ttnpb.ApplicationIdentifiers,
// ttnpb.ClientIdentifiers, ttnpb.EndDeviceIdentifiers,
// ttnpb.GatewayIdentifiers, ttnpb.OrganizationIdentifiers or
// ttnpb.UserIdentifiers.
// The reason for panicking is that taking the unique identifier of a nil or
// zero value may result in unexpected and potentially harmful behavior.
func ID(ctx context.Context, id ttnpb.Identifiers) (res string) {
	switch val := id.(type) {
	case ttnpb.ApplicationIdentifiers:
		res = val.ApplicationID
	case *ttnpb.ApplicationIdentifiers:
		res = val.ApplicationID
	case ttnpb.ClientIdentifiers:
		res = val.ClientID
	case *ttnpb.ClientIdentifiers:
		res = val.ClientID
	case ttnpb.EndDeviceIdentifiers:
		if val.ApplicationID != "" && val.DeviceID != "" {
			res = fmt.Sprintf("%v%v%v", val.ApplicationID, separator, val.DeviceID)
		}
	case *ttnpb.EndDeviceIdentifiers:
		if val.ApplicationID != "" && val.DeviceID != "" {
			res = fmt.Sprintf("%v%v%v", val.ApplicationID, separator, val.DeviceID)
		}
	case ttnpb.GatewayIdentifiers:
		res = val.GatewayID
	case *ttnpb.GatewayIdentifiers:
		res = val.GatewayID
	case ttnpb.OrganizationIdentifiers:
		res = val.OrganizationID
	case *ttnpb.OrganizationIdentifiers:
		res = val.OrganizationID
	case ttnpb.UserIdentifiers:
		res = val.UserID
	case *ttnpb.UserIdentifiers:
		res = val.UserID
	case ttnpb.EntityIdentifiers:
		return ID(ctx, val.Identifiers())
	case *ttnpb.EntityIdentifiers:
		return ID(ctx, val.Identifiers())
	default:
		panic(fmt.Errorf("failed to determine unique ID: %T is not a valid ttnpb.Identifiers", id))
	}
	if res == "" {
		panic(fmt.Errorf("failed to determine unique ID: the primary identifier is empty"))
	}
	return
}

// WithContext returns the given context.
func WithContext(ctx context.Context, uid string) (context.Context, error) {
	return ctx, nil
}

// ToApplicationID returns the application identifier of the specified unique ID.
func ToApplicationID(uid string) (ttnpb.ApplicationIdentifiers, error) {
	ids := ttnpb.ApplicationIdentifiers{ApplicationID: uid}
	if err := ids.ValidateFields("application_id"); err != nil {
		return ttnpb.ApplicationIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return ids, nil
}

// ToClientID returns the client identifier of the specified unique ID.
func ToClientID(uid string) (ttnpb.ClientIdentifiers, error) {
	ids := ttnpb.ClientIdentifiers{ClientID: uid}
	if err := ids.ValidateFields("client_id"); err != nil {
		return ttnpb.ClientIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return ids, nil
}

// ToDeviceID returns the end device identifier of the specified unique ID.
func ToDeviceID(uid string) (id ttnpb.EndDeviceIdentifiers, err error) {
	if parts := strings.SplitN(uid, separator, 2); len(parts) == 2 {
		devIDs := ttnpb.EndDeviceIdentifiers{DeviceID: parts[1], ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: parts[0]}}
		if err := devIDs.ValidateFields("device_id", "application_ids"); err != nil {
			return ttnpb.EndDeviceIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
		}
		return devIDs, nil
	}
	return ttnpb.EndDeviceIdentifiers{}, errFormat.WithAttributes("uid", uid)
}

// ToGatewayID returns the gateway identifier of the specified unique ID.
func ToGatewayID(uid string) (ttnpb.GatewayIdentifiers, error) {
	ids := ttnpb.GatewayIdentifiers{GatewayID: uid}
	if err := ids.ValidateFields("gateway_id"); err != nil {
		return ttnpb.GatewayIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return ids, nil
}

// ToOrganizationID returns the organization identifier of the specified unique ID.
func ToOrganizationID(uid string) (ttnpb.OrganizationIdentifiers, error) {
	ids := ttnpb.OrganizationIdentifiers{OrganizationID: uid}
	if err := ids.ValidateFields("organization_id"); err != nil {
		return ttnpb.OrganizationIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return ids, nil
}

// ToUserID returns the user identifier of the specified unique ID.
func ToUserID(uid string) (ttnpb.UserIdentifiers, error) {
	ids := ttnpb.UserIdentifiers{UserID: uid}
	if err := ids.ValidateFields("user_id"); err != nil {
		return ttnpb.UserIdentifiers{}, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return ids, nil
}

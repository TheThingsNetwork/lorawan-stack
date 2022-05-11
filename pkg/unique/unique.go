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

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	errUniqueIdentifier = errors.DefineInvalidArgument("unique_identifier", "invalid unique identifier `{uid}`")
	errFormat           = errors.DefineInvalidArgument("format", "invalid format in value `{value}`")
)

// ID returns the unique identifier of the specified identifiers.
// This function panics if the resulting identifier is invalid.
// The reason for panicking is that taking the unique identifier of a nil or
// zero value may result in unexpected and potentially harmful behavior.
func ID(ctx context.Context, id ttnpb.IDStringer) (res string) {
	res = id.IDString()
	if res == "" || strings.HasPrefix(res, ".") || strings.HasSuffix(res, ".") {
		panic(fmt.Errorf("failed to determine unique ID: the primary identifier is invalid"))
	}
	return res
}

// GenericID returns a generic selector identifier for an entity.
func GenericID(ctx context.Context, uidPattern string) (res string) {
	return uidPattern
}

// WithContext returns the given context.
func WithContext(ctx context.Context, uid string) (context.Context, error) {
	return ctx, nil
}

// ToApplicationID returns the application identifier of the specified unique ID.
func ToApplicationID(uid string) (id *ttnpb.ApplicationIdentifiers, err error) {
	id = &ttnpb.ApplicationIdentifiers{ApplicationId: uid}
	if err := id.ValidateFields("application_id"); err != nil {
		return nil, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

// ToClientID returns the client identifier of the specified unique ID.
func ToClientID(uid string) (id *ttnpb.ClientIdentifiers, err error) {
	id = &ttnpb.ClientIdentifiers{ClientId: uid}
	if err := id.ValidateFields("client_id"); err != nil {
		return nil, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

// ToDeviceID returns the end device identifier of the specified unique ID.
func ToDeviceID(uid string) (id *ttnpb.EndDeviceIdentifiers, err error) {
	id = &ttnpb.EndDeviceIdentifiers{}
	sepIdx := strings.Index(uid, ".")
	if sepIdx == -1 {
		return nil, errFormat.WithAttributes("value", uid)
	}
	id.ApplicationIds = &ttnpb.ApplicationIdentifiers{
		ApplicationId: uid[:sepIdx],
	}
	id.DeviceId = uid[sepIdx+1:]
	if err := id.ValidateFields("device_id", "application_ids"); err != nil {
		return nil, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

// ToGatewayID returns the gateway identifier of the specified unique ID.
func ToGatewayID(uid string) (id *ttnpb.GatewayIdentifiers, err error) {
	id = &ttnpb.GatewayIdentifiers{GatewayId: uid}
	if err := id.ValidateFields("gateway_id"); err != nil {
		return nil, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

// ToOrganizationID returns the organization identifier of the specified unique ID.
func ToOrganizationID(uid string) (id *ttnpb.OrganizationIdentifiers, err error) {
	id = &ttnpb.OrganizationIdentifiers{OrganizationId: uid}
	if err := id.ValidateFields("organization_id"); err != nil {
		return nil, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

// ToUserID returns the user identifier of the specified unique ID.
func ToUserID(uid string) (id *ttnpb.UserIdentifiers, err error) {
	id = &ttnpb.UserIdentifiers{UserId: uid}
	if err := id.ValidateFields("user_id"); err != nil {
		return nil, errUniqueIdentifier.WithCause(err).WithAttributes("uid", uid)
	}
	return id, nil
}

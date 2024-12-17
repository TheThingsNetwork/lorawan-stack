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

package ttnpb

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	types "go.thethings.network/lorawan-stack/v3/pkg/types"
)

// IsZero returns true if all identifiers have zero-values.
func (ids *ApplicationIdentifiers) IsZero() bool {
	return ids == nil || ids.ApplicationId == ""
}

// FieldIsZero returns whether path p is zero.
func (v *ApplicationIdentifiers) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "application_id":
		return v.ApplicationId == ""
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// IsZero returns true if all identifiers have zero-values.
func (ids *ClientIdentifiers) IsZero() bool {
	return ids == nil || ids.ClientId == ""
}

// IsZero reports whether ids represent zero identifiers.
func (ids *EndDeviceIdentifiers) IsZero() bool {
	if ids == nil {
		return true
	}
	return ids.DeviceId == "" &&
		ids.ApplicationIds == nil &&
		types.MustDevAddr(ids.DevAddr).OrZero().IsZero() &&
		types.MustEUI64(ids.DevEui).OrZero().IsZero() &&
		ids.JoinEui == nil
}

// FieldIsZero returns whether path p is zero.
func (v *EndDeviceIdentifiers) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "application_ids":
		return v.ApplicationIds == nil
	case "application_ids.application_id":
		return v.ApplicationIds.FieldIsZero("application_id")
	case "dev_addr":
		return v.DevAddr == nil
	case "dev_eui":
		return v.DevEui == nil
	case "device_id":
		return v.DeviceId == ""
	case "join_eui":
		return v.JoinEui == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// IsZero returns true if all identifiers have zero-values.
func (ids *GatewayIdentifiers) IsZero() bool {
	if ids == nil {
		return true
	}
	return ids.GatewayId == "" && ids.Eui == nil
}

// IsZero returns true if all identifiers have zero-values.
func (ids *OrganizationIdentifiers) IsZero() bool {
	return ids == nil || ids.OrganizationId == ""
}

// IsZero returns true if all identifiers have zero-values.
func (ids *UserIdentifiers) IsZero() bool {
	if ids == nil {
		return true
	}
	return ids.GetUserId() == "" && ids.GetEmail() == ""
}

// IsZero returns true if all identifiers have zero-values.
func (ids *MACSettingsProfileIdentifiers) IsZero() bool {
	if ids == nil {
		return true
	}
	return ids.ProfileId == "" && ids.ApplicationIds == nil
}

// GetOrganizationOrUserIdentifiers returns the OrganizationIdentifiers as *OrganizationOrUserIdentifiers.
func (ids *OrganizationIdentifiers) GetOrganizationOrUserIdentifiers() *OrganizationOrUserIdentifiers {
	if ids == nil {
		return nil
	}
	return &OrganizationOrUserIdentifiers{Ids: &OrganizationOrUserIdentifiers_OrganizationIds{
		OrganizationIds: ids,
	}}
}

// GetOrganizationOrUserIdentifiers returns the UserIdentifiers as *OrganizationOrUserIdentifiers.
func (ids *UserIdentifiers) GetOrganizationOrUserIdentifiers() *OrganizationOrUserIdentifiers {
	if ids == nil {
		return nil
	}
	return &OrganizationOrUserIdentifiers{Ids: &OrganizationOrUserIdentifiers_UserIds{
		UserIds: ids,
	}}
}

var errIdentifiers = errors.DefineInvalidArgument("identifiers", "invalid identifiers")

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (ids *EndDeviceIdentifiers) ValidateContext(context.Context) error {
	if err := ids.ValidateFields(); err != nil {
		return errIdentifiers.WithCause(err)
	}
	return nil
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (ids *ApplicationIdentifiers) ValidateContext(context.Context) error {
	if err := ids.ValidateFields(); err != nil {
		return errIdentifiers.WithCause(err)
	}
	return nil
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (ids *GatewayIdentifiers) ValidateContext(context.Context) error {
	if err := ids.ValidateFields(); err != nil {
		return errIdentifiers.WithCause(err)
	}
	return nil
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (ids *UserIdentifiers) ValidateContext(context.Context) error {
	if err := ids.ValidateFields(); err != nil {
		return errIdentifiers.WithCause(err)
	}
	return nil
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (ids *MACSettingsProfileIdentifiers) ValidateContext(context.Context) error {
	if err := ids.ValidateFields(); err != nil {
		return errIdentifiers.WithCause(err)
	}
	return nil
}

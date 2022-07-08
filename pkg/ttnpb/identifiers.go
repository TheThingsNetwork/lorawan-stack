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
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// IsZero returns true if all identifiers have zero-values.
func (ids ApplicationIdentifiers) IsZero() bool {
	return ids.ApplicationId == ""
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
func (ids ClientIdentifiers) IsZero() bool {
	return ids.ClientId == ""
}

// IsZero reports whether ids represent zero identifiers.
func (ids EndDeviceIdentifiers) IsZero() bool {
	return ids.DeviceId == "" &&
		ids.ApplicationIds == nil &&
		(ids.DevAddr == nil || ids.DevAddr.IsZero()) &&
		(ids.DevEui == nil || ids.DevEui.IsZero()) &&
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
func (ids GatewayIdentifiers) IsZero() bool {
	return ids.GatewayId == "" && ids.Eui == nil
}

// IsZero returns true if all identifiers have zero-values.
func (ids OrganizationIdentifiers) IsZero() bool {
	return ids.OrganizationId == ""
}

// IsZero returns true if all identifiers have zero-values.
func (ids UserIdentifiers) IsZero() bool {
	return ids.GetUserId() == "" && ids.GetEmail() == ""
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

// Copy stores a copy of ids in x and returns it.
func (ids EndDeviceIdentifiers) Copy(x *EndDeviceIdentifiers) *EndDeviceIdentifiers {
	*x = EndDeviceIdentifiers{
		DeviceId:      ids.DeviceId,
		XXX_sizecache: ids.XXX_sizecache,
	}
	if ids.ApplicationIds != nil {
		x.ApplicationIds = &ApplicationIdentifiers{
			ApplicationId: ids.GetApplicationIds().GetApplicationId(),
		}
	}
	if ids.DevEui != nil {
		x.DevEui = ids.DevEui.Copy(&types.EUI64{})
	}
	if ids.JoinEui != nil {
		x.JoinEui = ids.JoinEui.Copy(&types.EUI64{})
	}
	if ids.DevAddr != nil {
		x.DevAddr = ids.DevAddr.Copy(&types.DevAddr{})
	}
	return x
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

func (ids *GatewayIdentifiers) GetEui() *types.EUI64 {
	if ids == nil {
		return nil
	}
	return ids.Eui
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (ids *UserIdentifiers) ValidateContext(context.Context) error {
	if err := ids.ValidateFields(); err != nil {
		return errIdentifiers.WithCause(err)
	}
	return nil
}

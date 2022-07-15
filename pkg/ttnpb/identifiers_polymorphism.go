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

import "fmt"

// IDStringer identifies the string type and ID of identifiers.
type IDStringer interface {
	EntityType() string
	IDString() string
}

// GetEntityIdentifiers returns the ApplicationIdentifiers as EntityIdentifiers.
func (ids *ApplicationIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_ApplicationIds{
		ApplicationIds: ids,
	}}
}

// GetEntityIdentifiers returns the ClientIdentifiers as EntityIdentifiers.
func (ids *ClientIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_ClientIds{
		ClientIds: ids,
	}}
}

// GetEntityIdentifiers returns the EndDeviceIdentifiers as EntityIdentifiers.
func (ids *EndDeviceIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_DeviceIds{
		DeviceIds: ids,
	}}
}

// GetEntityIdentifiers returns the GatewayIdentifiers as EntityIdentifiers.
func (ids *GatewayIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_GatewayIds{
		GatewayIds: ids,
	}}
}

// GetEntityIdentifiers returns the OrganizationIdentifiers as EntityIdentifiers.
func (ids *OrganizationIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_OrganizationIds{
		OrganizationIds: ids,
	}}
}

// GetEntityIdentifiers returns the UserIdentifiers as EntityIdentifiers.
func (ids *UserIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_UserIds{
		UserIds: ids,
	}}
}

// GetEntityIdentifiers returns itself.
func (ids *EntityIdentifiers) GetEntityIdentifiers() *EntityIdentifiers { return ids }

// GetEntityIdentifiers returns the Identifiers inside the oneof as EntityIdentifiers.
func (ids *OrganizationOrUserIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	switch oneof := ids.Ids.(type) {
	case nil:
		return nil
	case *OrganizationOrUserIdentifiers_OrganizationIds:
		return oneof.OrganizationIds.GetEntityIdentifiers()
	case *OrganizationOrUserIdentifiers_UserIds:
		return oneof.UserIds.GetEntityIdentifiers()
	default:
		panic("missed oneof type in OrganizationOrUserIdentifiers.GetEntityIdentifiers()")
	}
}

// IDString returns the ID string of this Identifier.
func (ids *ApplicationIdentifiers) IDString() string { return ids.GetApplicationId() }

// IDString returns the ID string of this Identifier.
func (ids *ClientIdentifiers) IDString() string { return ids.GetClientId() }

// IDString returns the ID string of this Identifier.
func (ids *EndDeviceIdentifiers) IDString() string {
	if ids == nil {
		return ""
	}
	return fmt.Sprintf("%s.%s", ids.GetApplicationIds().IDString(), ids.GetDeviceId())
}

// IDString returns the ID string of this Identifier.
func (ids *GatewayIdentifiers) IDString() string { return ids.GetGatewayId() }

// IDString returns the ID string of this Identifier.
func (ids *OrganizationIdentifiers) IDString() string { return ids.GetOrganizationId() }

// IDString returns the ID string of this Identifier.
func (ids *UserIdentifiers) IDString() string { return ids.GetUserId() }

// IDString returns the ID string of the Identifiers inside the oneof.
func (ids *EntityIdentifiers) IDString() string {
	if ids == nil {
		return ""
	}
	switch oneof := ids.Ids.(type) {
	case nil:
		return ""
	case *EntityIdentifiers_ApplicationIds:
		return oneof.ApplicationIds.IDString()
	case *EntityIdentifiers_ClientIds:
		return oneof.ClientIds.IDString()
	case *EntityIdentifiers_DeviceIds:
		return oneof.DeviceIds.IDString()
	case *EntityIdentifiers_GatewayIds:
		return oneof.GatewayIds.IDString()
	case *EntityIdentifiers_OrganizationIds:
		return oneof.OrganizationIds.IDString()
	case *EntityIdentifiers_UserIds:
		return oneof.UserIds.IDString()
	default:
		panic("missed oneof type in EntityIdentifiers.IDString()")
	}
}

// IDString returns the ID string of the Identifiers inside the oneof.
func (ids *OrganizationOrUserIdentifiers) IDString() string {
	if ids == nil {
		return ""
	}
	switch oneof := ids.Ids.(type) {
	case nil:
		return ""
	case *OrganizationOrUserIdentifiers_OrganizationIds:
		return oneof.OrganizationIds.IDString()
	case *OrganizationOrUserIdentifiers_UserIds:
		return oneof.UserIds.IDString()
	default:
		panic("missed oneof type in OrganizationOrUserIdentifiers.IDString()")
	}
}

// EntityType returns the entity type for this ID (application).
func (*ApplicationIdentifiers) EntityType() string { return "application" }

// EntityType returns the entity type for this ID (client).
func (*ClientIdentifiers) EntityType() string { return "client" }

// EntityType returns the entity type for this ID (end device).
func (*EndDeviceIdentifiers) EntityType() string { return "end device" }

// EntityType returns the entity type for this ID (gateway).
func (*GatewayIdentifiers) EntityType() string { return "gateway" }

// EntityType returns the entity type for this ID (organization).
func (*OrganizationIdentifiers) EntityType() string { return "organization" }

// EntityType returns the entity type for this ID (user).
func (*UserIdentifiers) EntityType() string { return "user" }

// EntityType returns the entity type for the Identifiers inside the oneof.
func (ids *EntityIdentifiers) EntityType() string {
	if ids == nil {
		return ""
	}
	switch oneof := ids.Ids.(type) {
	case nil:
		return ""
	case *EntityIdentifiers_ApplicationIds:
		return oneof.ApplicationIds.EntityType()
	case *EntityIdentifiers_ClientIds:
		return oneof.ClientIds.EntityType()
	case *EntityIdentifiers_DeviceIds:
		return oneof.DeviceIds.EntityType()
	case *EntityIdentifiers_GatewayIds:
		return oneof.GatewayIds.EntityType()
	case *EntityIdentifiers_OrganizationIds:
		return oneof.OrganizationIds.EntityType()
	case *EntityIdentifiers_UserIds:
		return oneof.UserIds.EntityType()
	default:
		panic("missed oneof type in EntityIdentifiers.EntityType()")
	}
}

// EntityType returns the entity type for the Identifiers inside the oneof.
func (ids *OrganizationOrUserIdentifiers) EntityType() string {
	if ids == nil {
		return ""
	}
	switch oneof := ids.Ids.(type) {
	case nil:
		return ""
	case *OrganizationOrUserIdentifiers_OrganizationIds:
		return oneof.OrganizationIds.EntityType()
	case *OrganizationOrUserIdentifiers_UserIds:
		return oneof.UserIds.EntityType()
	default:
		panic("missed oneof type in OrganizationOrUserIdentifiers.EntityType()")
	}
}

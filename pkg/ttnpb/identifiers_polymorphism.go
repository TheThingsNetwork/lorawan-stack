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
	return &EntityIdentifiers{Ids: &EntityIdentifiers_ApplicationIDs{
		ApplicationIDs: ids,
	}}
}

// GetEntityIdentifiers returns the ClientIdentifiers as EntityIdentifiers.
func (ids *ClientIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_ClientIDs{
		ClientIDs: ids,
	}}
}

// GetEntityIdentifiers returns the EndDeviceIdentifiers as EntityIdentifiers.
func (ids *EndDeviceIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_DeviceIDs{
		DeviceIDs: ids,
	}}
}

// GetEntityIdentifiers returns the GatewayIdentifiers as EntityIdentifiers.
func (ids *GatewayIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_GatewayIDs{
		GatewayIDs: ids,
	}}
}

// GetEntityIdentifiers returns the OrganizationIdentifiers as EntityIdentifiers.
func (ids *OrganizationIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_OrganizationIDs{
		OrganizationIDs: ids,
	}}
}

// GetEntityIdentifiers returns the UserIdentifiers as EntityIdentifiers.
func (ids *UserIdentifiers) GetEntityIdentifiers() *EntityIdentifiers {
	if ids == nil {
		return nil
	}
	return &EntityIdentifiers{Ids: &EntityIdentifiers_UserIDs{
		UserIDs: ids,
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
	case *OrganizationOrUserIdentifiers_OrganizationIDs:
		return oneof.OrganizationIDs.GetEntityIdentifiers()
	case *OrganizationOrUserIdentifiers_UserIDs:
		return oneof.UserIDs.GetEntityIdentifiers()
	default:
		panic("missed oneof type in OrganizationOrUserIdentifiers.GetEntityIdentifiers()")
	}
}

// IDString returns the ID string of this Identifier.
func (ids ApplicationIdentifiers) IDString() string { return ids.ApplicationID }

// IDString returns the ID string of this Identifier.
func (ids ClientIdentifiers) IDString() string { return ids.ClientID }

// IDString returns the ID string of this Identifier.
func (ids EndDeviceIdentifiers) IDString() string {
	return fmt.Sprintf("%s.%s", ids.ApplicationIdentifiers.IDString(), ids.DeviceID)
}

// IDString returns the ID string of this Identifier.
func (ids GatewayIdentifiers) IDString() string { return ids.GatewayID }

// IDString returns the ID string of this Identifier.
func (ids OrganizationIdentifiers) IDString() string { return ids.OrganizationID }

// IDString returns the ID string of this Identifier.
func (ids UserIdentifiers) IDString() string { return ids.UserID }

// IDString returns the ID string of the Identifiers inside the oneof.
func (ids EntityIdentifiers) IDString() string {
	switch oneof := ids.Ids.(type) {
	case *EntityIdentifiers_ApplicationIDs:
		return oneof.ApplicationIDs.IDString()
	case *EntityIdentifiers_ClientIDs:
		return oneof.ClientIDs.IDString()
	case *EntityIdentifiers_DeviceIDs:
		return oneof.DeviceIDs.IDString()
	case *EntityIdentifiers_GatewayIDs:
		return oneof.GatewayIDs.IDString()
	case *EntityIdentifiers_OrganizationIDs:
		return oneof.OrganizationIDs.IDString()
	case *EntityIdentifiers_UserIDs:
		return oneof.UserIDs.IDString()
	default:
		panic("missed oneof type in EntityIdentifiers.IDString()")
	}
}

// IDString returns the ID string of the Identifiers inside the oneof.
func (ids OrganizationOrUserIdentifiers) IDString() string {
	switch oneof := ids.Ids.(type) {
	case *OrganizationOrUserIdentifiers_OrganizationIDs:
		return oneof.OrganizationIDs.IDString()
	case *OrganizationOrUserIdentifiers_UserIDs:
		return oneof.UserIDs.IDString()
	default:
		panic("missed oneof type in OrganizationOrUserIdentifiers.IDString()")
	}
}

// EntityType returns the entity type for this ID (application).
func (ApplicationIdentifiers) EntityType() string { return "application" }

// EntityType returns the entity type for this ID (client).
func (ClientIdentifiers) EntityType() string { return "client" }

// EntityType returns the entity type for this ID (end device).
func (EndDeviceIdentifiers) EntityType() string { return "end device" }

// EntityType returns the entity type for this ID (gateway).
func (GatewayIdentifiers) EntityType() string { return "gateway" }

// EntityType returns the entity type for this ID (organization).
func (OrganizationIdentifiers) EntityType() string { return "organization" }

// EntityType returns the entity type for this ID (user).
func (UserIdentifiers) EntityType() string { return "user" }

// EntityType returns the entity type for the Identifiers inside the oneof.
func (ids EntityIdentifiers) EntityType() string {
	switch oneof := ids.Ids.(type) {
	case *EntityIdentifiers_ApplicationIDs:
		return oneof.ApplicationIDs.EntityType()
	case *EntityIdentifiers_ClientIDs:
		return oneof.ClientIDs.EntityType()
	case *EntityIdentifiers_DeviceIDs:
		return oneof.DeviceIDs.EntityType()
	case *EntityIdentifiers_GatewayIDs:
		return oneof.GatewayIDs.EntityType()
	case *EntityIdentifiers_OrganizationIDs:
		return oneof.OrganizationIDs.EntityType()
	case *EntityIdentifiers_UserIDs:
		return oneof.UserIDs.EntityType()
	default:
		panic("missed oneof type in EntityIdentifiers.EntityType()")
	}
}

// EntityType returns the entity type for the Identifiers inside the oneof.
func (ids OrganizationOrUserIdentifiers) EntityType() string {
	switch oneof := ids.Ids.(type) {
	case *OrganizationOrUserIdentifiers_OrganizationIDs:
		return oneof.OrganizationIDs.EntityType()
	case *OrganizationOrUserIdentifiers_UserIDs:
		return oneof.UserIDs.EntityType()
	default:
		panic("missed oneof type in OrganizationOrUserIdentifiers.EntityType()")
	}
}

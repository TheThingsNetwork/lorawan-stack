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

// Identifiers is the interface implemented by all (single) identifiers.
type Identifiers interface {
	EntityType() string
	IDString() string
	Identifiers() Identifiers
	EntityIdentifiers() *EntityIdentifiers
	CombinedIdentifiers() *CombinedIdentifiers
}

// EntityIdentifiers returns the ApplicationIdentifiers as EntityIdentifiers.
func (ids ApplicationIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_ApplicationIDs{
		ApplicationIDs: &ids,
	}}
}

// EntityIdentifiers returns the ClientIdentifiers as EntityIdentifiers.
func (ids ClientIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_ClientIDs{
		ClientIDs: &ids,
	}}
}

// EntityIdentifiers returns the EndDeviceIdentifiers as EntityIdentifiers.
func (ids EndDeviceIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_DeviceIDs{
		DeviceIDs: &ids,
	}}
}

// EntityIdentifiers returns the GatewayIdentifiers as EntityIdentifiers.
func (ids GatewayIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_GatewayIDs{
		GatewayIDs: &ids,
	}}
}

// EntityIdentifiers implements returns theOrganizationIdentifiers as EntityIdentifiers.
func (ids OrganizationIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_OrganizationIDs{
		OrganizationIDs: &ids,
	}}
}

// EntityIdentifiers returns the UserIdentifiers as EntityIdentifiers.
func (ids UserIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_UserIDs{
		UserIDs: &ids,
	}}
}

// EntityIdentifiers returns itself.
func (ids EntityIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &ids
}

// EntityIdentifiers returns the Identifiers inside the oneof as EntityIdentifiers.
func (ids OrganizationOrUserIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	switch oneof := ids.Ids.(type) {
	case *OrganizationOrUserIdentifiers_OrganizationIDs:
		return oneof.OrganizationIDs.EntityIdentifiers()
	case *OrganizationOrUserIdentifiers_UserIDs:
		return oneof.UserIDs.EntityIdentifiers()
	default:
		panic("missed oneof type in OrganizationOrUserIdentifiers.EntityIdentifiers()")
	}
}

// Identifiers returns itself.
func (ids ApplicationIdentifiers) Identifiers() Identifiers { return &ids }

// Identifiers returns itself.
func (ids ClientIdentifiers) Identifiers() Identifiers { return &ids }

// Identifiers returns itself.
func (ids EndDeviceIdentifiers) Identifiers() Identifiers { return &ids }

// Identifiers returns itself.
func (ids GatewayIdentifiers) Identifiers() Identifiers { return &ids }

// Identifiers returns itself.
func (ids OrganizationIdentifiers) Identifiers() Identifiers { return &ids }

// Identifiers returns itself.
func (ids UserIdentifiers) Identifiers() Identifiers { return &ids }

// Identifiers returns the concrete identifiers inside the oneof.
func (ids EntityIdentifiers) Identifiers() Identifiers {
	switch oneof := ids.Ids.(type) {
	case *EntityIdentifiers_ApplicationIDs:
		return oneof.ApplicationIDs
	case *EntityIdentifiers_ClientIDs:
		return oneof.ClientIDs
	case *EntityIdentifiers_DeviceIDs:
		return oneof.DeviceIDs
	case *EntityIdentifiers_GatewayIDs:
		return oneof.GatewayIDs
	case *EntityIdentifiers_OrganizationIDs:
		return oneof.OrganizationIDs
	case *EntityIdentifiers_UserIDs:
		return oneof.UserIDs
	default:
		panic("missed oneof type in EntityIdentifiers.Identifiers()")
	}
}

// Identifiers returns the concrete identifiers inside the oneof.
func (ids OrganizationOrUserIdentifiers) Identifiers() Identifiers {
	switch oneof := ids.Ids.(type) {
	case *OrganizationOrUserIdentifiers_OrganizationIDs:
		return oneof.OrganizationIDs
	case *OrganizationOrUserIdentifiers_UserIDs:
		return oneof.UserIDs
	default:
		panic("missed oneof type in OrganizationOrUserIdentifiers.Identifiers()")
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

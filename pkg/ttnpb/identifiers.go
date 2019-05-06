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

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// IsZero returns true if all identifiers have zero-values.
func (ids ApplicationIdentifiers) IsZero() bool {
	return ids.ApplicationID == ""
}

// IsZero returns true if all identifiers have zero-values.
func (ids ClientIdentifiers) IsZero() bool {
	return ids.ClientID == ""
}

// IsZero reports whether ids represent zero identifiers.
func (ids EndDeviceIdentifiers) IsZero() bool {
	return ids.GetDeviceID() == "" &&
		ids.GetApplicationID() == "" &&
		(ids.DevAddr == nil || ids.DevAddr.IsZero()) &&
		(ids.DevEUI == nil || ids.DevEUI.IsZero()) &&
		(ids.JoinEUI == nil || ids.JoinEUI.IsZero())
}

// IsZero returns true if all identifiers have zero-values.
func (ids GatewayIdentifiers) IsZero() bool {
	return ids.GatewayID == "" && ids.EUI == nil
}

// IsZero returns true if all identifiers have zero-values.
func (ids OrganizationIdentifiers) IsZero() bool {
	return ids.OrganizationID == ""
}

// IsZero returns true if all identifiers have zero-values.
func (ids UserIdentifiers) IsZero() bool {
	return ids.UserID == "" && ids.Email == ""
}

// Identifiers interface for Entity Identifiers.
type Identifiers interface {
	CombinedIdentifiers() *CombinedIdentifiers
}

// CombinedIdentifiers returns the EntityIdentifiers as a CombinedIdentifiers type.
func (ids EntityIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return &CombinedIdentifiers{EntityIdentifiers: []*EntityIdentifiers{&ids}}
}

// Identifiers returns the actual Identifiers inside the oneof.
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

// IDString returns the ID string of the Identifiers inside the oneof.
func (ids EntityIdentifiers) IDString() string {
	switch oneof := ids.Ids.(type) {
	case *EntityIdentifiers_ApplicationIDs:
		return oneof.ApplicationIDs.GetApplicationID()
	case *EntityIdentifiers_ClientIDs:
		return oneof.ClientIDs.GetClientID()
	case *EntityIdentifiers_DeviceIDs:
		return fmt.Sprintf("%s.%s", oneof.DeviceIDs.GetApplicationID(), oneof.DeviceIDs.GetDeviceID())
	case *EntityIdentifiers_GatewayIDs:
		return oneof.GatewayIDs.GetGatewayID()
	case *EntityIdentifiers_OrganizationIDs:
		return oneof.OrganizationIDs.GetOrganizationID()
	case *EntityIdentifiers_UserIDs:
		return oneof.UserIDs.GetUserID()
	default:
		panic("missed oneof type in EntityIdentifiers.IDString()")
	}
}

// EntityType returns the entity type for the Identifiers inside the oneof.
func (ids EntityIdentifiers) EntityType() string {
	switch ids.Ids.(type) {
	case *EntityIdentifiers_ApplicationIDs:
		return "application"
	case *EntityIdentifiers_ClientIDs:
		return "client"
	case *EntityIdentifiers_DeviceIDs:
		return "end device"
	case *EntityIdentifiers_GatewayIDs:
		return "gateway"
	case *EntityIdentifiers_OrganizationIDs:
		return "organization"
	case *EntityIdentifiers_UserIDs:
		return "user"
	default:
		panic("missed oneof type in EntityIdentifiers.EntityType()")
	}
}

// EntityIdentifiers returns the ApplicationIdentifiers as EntityIdentifiers.
func (ids ApplicationIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_ApplicationIDs{
		ApplicationIDs: &ids,
	}}
}

// CombinedIdentifiers returns the ApplicationIdentifiers as CombinedIdentifiers.
func (ids ApplicationIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// EntityIdentifiers returns the ClientIdentifiers as EntityIdentifiers.
func (ids ClientIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_ClientIDs{
		ClientIDs: &ids,
	}}
}

// CombinedIdentifiers returns the ClientIdentifiers as CombinedIdentifiers.
func (ids ClientIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// EntityIdentifiers returns the EndDeviceIdentifiers as EntityIdentifiers.
func (ids EndDeviceIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_DeviceIDs{
		DeviceIDs: &ids,
	}}
}

// CombinedIdentifiers returns the EndDeviceIdentifiers as CombinedIdentifiers.
func (ids EndDeviceIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// EntityIdentifiers returns the GatewayIdentifiers as EntityIdentifiers.
func (ids GatewayIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_GatewayIDs{
		GatewayIDs: &ids,
	}}
}

// CombinedIdentifiers returns the GatewayIdentifiers as CombinedIdentifiers.
func (ids GatewayIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// EntityIdentifiers implements returns theOrganizationIdentifiers as EntityIdentifiers.
func (ids OrganizationIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_OrganizationIDs{
		OrganizationIDs: &ids,
	}}
}

// CombinedIdentifiers returns the OrganizationIdentifiers as CombinedIdentifiers.
func (ids OrganizationIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// OrganizationOrUserIdentifiers returns the OrganizationIdentifiers as *OrganizationOrUserIdentifiers.
func (ids OrganizationIdentifiers) OrganizationOrUserIdentifiers() *OrganizationOrUserIdentifiers {
	return &OrganizationOrUserIdentifiers{Ids: &OrganizationOrUserIdentifiers_OrganizationIDs{
		OrganizationIDs: &ids,
	}}
}

// EntityIdentifiers returns the UserIdentifiers as EntityIdentifiers.
func (ids UserIdentifiers) EntityIdentifiers() *EntityIdentifiers {
	return &EntityIdentifiers{Ids: &EntityIdentifiers_UserIDs{
		UserIDs: &ids,
	}}
}

// CombinedIdentifiers returns the UserIdentifiers as CombinedIdentifiers.
func (ids UserIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// OrganizationOrUserIdentifiers returns the UserIdentifiers as *OrganizationOrUserIdentifiers.
func (ids UserIdentifiers) OrganizationOrUserIdentifiers() *OrganizationOrUserIdentifiers {
	return &OrganizationOrUserIdentifiers{Ids: &OrganizationOrUserIdentifiers_UserIDs{
		UserIDs: &ids,
	}}
}

// Identifiers returns the OrganizationOrUserIdentifiers as Identifiers.
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

// EntityIdentifiers returns the OrganizationOrUserIdentifiers as EntityIdentifiers.
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

// CombinedIdentifiers returns the OrganizationOrUserIdentifiers as CombinedIdentifiers.
func (ids OrganizationOrUserIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// CombineIdentifiers merges the identifiers of the multiple entities.
func CombineIdentifiers(ids ...Identifiers) *CombinedIdentifiers {
	combined := &CombinedIdentifiers{}
	for _, id := range ids {
		combined.EntityIdentifiers = append(combined.EntityIdentifiers, id.CombinedIdentifiers().GetEntityIdentifiers()...)
	}
	return combined
}

// CombinedIdentifiers implements Identifiers.
func (ids *CombinedIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids
}

// CombinedIdentifiers implements Identifiers.
func (m *ListApplicationsRequest) CombinedIdentifiers() *CombinedIdentifiers {
	if m.Collaborator != nil {
		return m.Collaborator.CombinedIdentifiers()
	}
	return &CombinedIdentifiers{}
}

// CombinedIdentifiers implements Identifiers.
func (m *ListClientsRequest) CombinedIdentifiers() *CombinedIdentifiers {
	if m.Collaborator != nil {
		return m.Collaborator.CombinedIdentifiers()
	}
	return &CombinedIdentifiers{}
}

// CombinedIdentifiers implements Identifiers.
func (m *ListGatewaysRequest) CombinedIdentifiers() *CombinedIdentifiers {
	if m.Collaborator != nil {
		return m.Collaborator.CombinedIdentifiers()
	}
	return &CombinedIdentifiers{}
}

// CombinedIdentifiers implements Identifiers.
func (m *ListOrganizationsRequest) CombinedIdentifiers() *CombinedIdentifiers {
	if m.Collaborator != nil {
		return m.Collaborator.CombinedIdentifiers()
	}
	return &CombinedIdentifiers{}
}

// CombinedIdentifiers implements Identifiers.
func (m *CreateApplicationRequest) CombinedIdentifiers() *CombinedIdentifiers {
	return m.Collaborator.CombinedIdentifiers()
}

// CombinedIdentifiers implements Identifiers.
func (m *CreateClientRequest) CombinedIdentifiers() *CombinedIdentifiers {
	return m.Collaborator.CombinedIdentifiers()
}

// CombinedIdentifiers implements Identifiers.
func (m *CreateGatewayRequest) CombinedIdentifiers() *CombinedIdentifiers {
	return m.Collaborator.CombinedIdentifiers()
}

// CombinedIdentifiers implements Identifiers.
func (m *CreateOrganizationRequest) CombinedIdentifiers() *CombinedIdentifiers {
	return m.Collaborator.CombinedIdentifiers()
}

// CombinedIdentifiers implements Identifiers.
func (m *DownlinkMessage) CombinedIdentifiers() *CombinedIdentifiers {
	return m.GetEndDeviceIDs().CombinedIdentifiers()
}

// CombinedIdentifiers implements Identifiers.
func (m *SetEndDeviceRequest) CombinedIdentifiers() *CombinedIdentifiers {
	return m.EndDevice.CombinedIdentifiers()
}

// CombinedIdentifiers implements Identifiers.
func (m *ListOAuthAccessTokensRequest) CombinedIdentifiers() *CombinedIdentifiers {
	return CombineIdentifiers(m.UserIDs, m.ClientIDs)
}

// CombinedIdentifiers implements Identifiers.
func (m *ListOAuthClientAuthorizationsRequest) CombinedIdentifiers() *CombinedIdentifiers {
	return m.UserIdentifiers.CombinedIdentifiers()
}

// CombinedIdentifiers implements Identifiers.
func (m *OAuthAccessTokenIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return CombineIdentifiers(m.UserIDs, m.ClientIDs)
}

// CombinedIdentifiers implements Identifiers.
func (m *OAuthClientAuthorizationIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return CombineIdentifiers(m.UserIDs, m.ClientIDs)
}

// CombinedIdentifiers implements Identifiers.
func (m *StreamEventsRequest) CombinedIdentifiers() *CombinedIdentifiers {
	return &CombinedIdentifiers{EntityIdentifiers: m.Identifiers}
}

// Copy stores a copy of ids in x and returns it.
func (ids EndDeviceIdentifiers) Copy(x *EndDeviceIdentifiers) *EndDeviceIdentifiers {
	*x = EndDeviceIdentifiers{
		DeviceID: ids.DeviceID,
		ApplicationIdentifiers: ApplicationIdentifiers{
			ApplicationID: ids.ApplicationID,
		},
		XXX_sizecache: ids.XXX_sizecache,
	}
	if ids.DevEUI != nil {
		x.DevEUI = ids.DevEUI.Copy(&types.EUI64{})
	}
	if ids.JoinEUI != nil {
		x.JoinEUI = ids.JoinEUI.Copy(&types.EUI64{})
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

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (ids *UserIdentifiers) ValidateContext(context.Context) error {
	if err := ids.ValidateFields(); err != nil {
		return errIdentifiers.WithCause(err)
	}
	return nil
}

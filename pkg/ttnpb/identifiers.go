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

package ttnpb

import (
	"encoding/gob"

	"go.thethings.network/lorawan-stack/pkg/types"
)

func init() {
	gob.Register(&OrganizationOrUserIdentifiers_UserID{})
	gob.Register(&OrganizationOrUserIdentifiers_OrganizationID{})
}

// IsZero returns true if all identifiers have zero-values.
func (ids UserIdentifiers) IsZero() bool {
	return ids.UserID == "" && ids.Email == ""
}

// Equals returns true if the receiver identifiers matches to other identifiers.
func (ids UserIdentifiers) Equals(other UserIdentifiers) bool {
	return ids.UserID == other.UserID && ids.Email == other.Email
}

// Contains returns true if other is contained in the receiver.
func (ids UserIdentifiers) Contains(other UserIdentifiers) bool {
	if other.IsZero() {
		return ids.IsZero()
	}

	return (other.UserID == "" || ids.UserID == other.UserID) &&
		(other.Email == "" || ids.Email == other.Email)
}

// Contains returns true if other is contained in the receiver.
func (ids ApplicationIdentifiers) Contains(other ApplicationIdentifiers) bool {
	return ids.ApplicationID == other.ApplicationID
}

// IsZero returns true if all identifiers have zero-values.
func (ids ApplicationIdentifiers) IsZero() bool {
	return ids.ApplicationID == ""
}

// GetEUI returns if set the EUI otherwise a zero-valued EUI.
func (ids GatewayIdentifiers) GetEUI() *types.EUI64 {
	if ids.EUI != nil {
		return ids.EUI
	}
	return new(types.EUI64)
}

// IsZero returns true if all identifiers have zero-values.
func (ids GatewayIdentifiers) IsZero() bool {
	return ids.GatewayID == "" && ids.EUI == nil
}

// Contains returns true if other is contained in the receiver.
func (ids GatewayIdentifiers) Contains(other GatewayIdentifiers) bool {
	if other.IsZero() {
		return ids.IsZero()
	}

	return (other.GatewayID == "" || (other.GatewayID != "" && ids.GatewayID == other.GatewayID)) &&
		((ids.EUI == nil && other.EUI == nil) || (ids.EUI != nil && other.EUI == nil) || ids.EUI.Equal(*other.EUI))
}

// IsZero returns true if all identifiers have zero-values.
func (ids ClientIdentifiers) IsZero() bool {
	return ids.ClientID == ""
}

// IsZero returns true if all identifiers have zero-values.
func (ids OrganizationIdentifiers) IsZero() bool {
	return ids.OrganizationID == ""
}

// Contains returns true if other is contained in the receiver.
func (ids OrganizationIdentifiers) Contains(other OrganizationIdentifiers) bool {
	return ids.OrganizationID == other.OrganizationID
}

// IsZero reports whether ids represent zero identifiers.
func (ids EndDeviceIdentifiers) IsZero() bool {
	return ids.GetDeviceID() == "" &&
		ids.GetApplicationID() == "" &&
		(ids.DevAddr == nil || ids.DevAddr.IsZero()) &&
		(ids.DevEUI == nil || ids.DevEUI.IsZero()) &&
		(ids.JoinEUI == nil || ids.JoinEUI.IsZero())
}

type Identifiers interface {
	CombinedIdentifiers() *CombinedIdentifiers
}

func (ids ApplicationIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return &CombinedIdentifiers{ApplicationIDs: []*ApplicationIdentifiers{&ids}}
}
func (ids ClientIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return &CombinedIdentifiers{ClientIDs: []*ClientIdentifiers{&ids}}
}
func (ids EndDeviceIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return &CombinedIdentifiers{DeviceIDs: []*EndDeviceIdentifiers{&ids}}
}
func (ids GatewayIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return &CombinedIdentifiers{GatewayIDs: []*GatewayIdentifiers{&ids}}
}
func (ids OrganizationIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return &CombinedIdentifiers{OrganizationIDs: []*OrganizationIdentifiers{&ids}}
}
func (ids UserIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return &CombinedIdentifiers{UserIDs: []*UserIdentifiers{&ids}}
}

func CombineIdentifiers(ids ...Identifiers) *CombinedIdentifiers {
	combined := &CombinedIdentifiers{}
	for _, id := range ids {
		asCombined := id.CombinedIdentifiers()
		combined.ApplicationIDs = append(combined.ApplicationIDs, asCombined.ApplicationIDs...)
		combined.ClientIDs = append(combined.ClientIDs, asCombined.ClientIDs...)
		combined.DeviceIDs = append(combined.DeviceIDs, asCombined.DeviceIDs...)
		combined.GatewayIDs = append(combined.GatewayIDs, asCombined.GatewayIDs...)
		combined.OrganizationIDs = append(combined.OrganizationIDs, asCombined.OrganizationIDs...)
		combined.UserIDs = append(combined.UserIDs, asCombined.UserIDs...)
	}
	return combined
}

func (ids *CombinedIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids
}

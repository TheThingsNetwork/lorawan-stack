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
	"github.com/TheThingsNetwork/ttn/pkg/types"
)

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

func (ids EndDeviceIdentifiers) IsZero() bool {
	return ids.GetDeviceID() == "" &&
		ids.GetApplicationID() == "" &&
		(ids.DevAddr == nil || ids.DevAddr.IsZero()) &&
		(ids.DevEUI == nil || ids.DevEUI.IsZero()) &&
		(ids.JoinEUI == nil || ids.JoinEUI.IsZero())
}

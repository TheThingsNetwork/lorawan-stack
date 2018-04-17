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

import "github.com/TheThingsNetwork/ttn/pkg/types"

// IsZero returns true if all identifiers have zero-values.
func (i UserIdentifiers) IsZero() bool {
	return i.UserID == "" && i.Email == ""
}

// Equals returns true if the receiver identifiers matches to other identifiers.
func (i UserIdentifiers) Equals(other UserIdentifiers) bool {
	return i.UserID == other.UserID && i.Email == other.Email
}

// Contains returns true if other is contained in the receiver.
func (i UserIdentifiers) Contains(other UserIdentifiers) bool {
	if other.IsZero() {
		return i.IsZero()
	}

	return (other.UserID == "" || i.UserID == other.UserID) &&
		(other.Email == "" || i.Email == other.Email)
}

// Contains returns true if other is contained in the receiver.
func (i ApplicationIdentifiers) Contains(other ApplicationIdentifiers) bool {
	return i.ApplicationID == other.ApplicationID
}

// IsZero returns true if all identifiers have zero-values.
func (i ApplicationIdentifiers) IsZero() bool {
	return i.ApplicationID == ""
}

// GetEUI returns if set the EUI otherwise a zero-valued EUI.
func (i GatewayIdentifiers) GetEUI() *types.EUI64 {
	if i.EUI != nil {
		return i.EUI
	}
	return new(types.EUI64)
}

// IsZero returns true if all identifiers have zero-values.
func (i GatewayIdentifiers) IsZero() bool {
	return i.GatewayID == "" && i.EUI == nil
}

// Contains returns true if other is contained in the receiver.
func (i GatewayIdentifiers) Contains(other GatewayIdentifiers) bool {
	if other.IsZero() {
		return i.IsZero()
	}

	return (other.GatewayID == "" || (other.GatewayID != "" && i.GatewayID == other.GatewayID)) &&
		((i.EUI == nil && other.EUI == nil) || (i.EUI != nil && other.EUI == nil) || i.EUI.Equal(*other.EUI))
}

// IsZero returns true if all identifiers have zero-values.
func (i ClientIdentifiers) IsZero() bool {
	return i.ClientID == ""
}

// IsZero returns true if all identifiers have zero-values.
func (i OrganizationIdentifiers) IsZero() bool {
	return i.OrganizationID == ""
}

// Contains returns true if other is contained in the receiver.
func (i OrganizationIdentifiers) Contains(other OrganizationIdentifiers) bool {
	return i.OrganizationID == other.OrganizationID
}

// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package store

import "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

// MembershipChain is a User -> (Membership -> Organization) -> Membership -> Entity chain.
type MembershipChain struct {
	UserIdentifiers         *ttnpb.UserIdentifiers
	RightsOnOrganization    *ttnpb.Rights
	OrganizationIdentifiers *ttnpb.OrganizationIdentifiers
	RightsOnEntity          *ttnpb.Rights
	EntityIdentifiers       *ttnpb.EntityIdentifiers
}

// GetRights returns the intersected rights.
func (m *MembershipChain) GetRights() *ttnpb.Rights {
	if m.RightsOnOrganization == nil {
		return m.RightsOnEntity.Implied()
	}
	return m.RightsOnEntity.Implied().Intersect(m.RightsOnOrganization.Implied())
}

// MembershipChains is a list of membership chains.
type MembershipChains []*MembershipChain

// GetRights returns the rights of the member on the entity.
func (c MembershipChains) GetRights(
	member *ttnpb.OrganizationOrUserIdentifiers, entityID ttnpb.IDStringer,
) *ttnpb.Rights {
	var entityRights *ttnpb.Rights
	for _, membership := range c {
		switch member.EntityType() {
		case "organization":
			if membership.OrganizationIdentifiers == nil ||
				membership.OrganizationIdentifiers.IDString() != member.IDString() {
				continue
			}
		case "user":
			if membership.UserIdentifiers == nil ||
				membership.UserIdentifiers.IDString() != member.IDString() {
				continue
			}
		default:
			continue
		}
		if membership.EntityIdentifiers.EntityType() != entityID.EntityType() ||
			membership.EntityIdentifiers.IDString() != entityID.IDString() {
			continue
		}
		entityRights = entityRights.Union(membership.GetRights())
	}
	return entityRights
}

// MemberByID defines a set containing a User or Organization Ids and their respective Rights.
type MemberByID struct {
	Ids    *ttnpb.OrganizationOrUserIdentifiers
	Rights *ttnpb.Rights
}

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

package store

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// Organization is the interface of all things that can be an organization. This
// can be used to build richer organization types that can still be read and written
// to a database.
type Organization interface {
	// GetOrganization returns the ttnpb.Organization that represents this organization.
	GetOrganization() *ttnpb.Organization
}

// OrganizationSpecializer returns a new Organization with the given base ttnpb.Organization.
type OrganizationSpecializer func(ttnpb.Organization) Organization

// OrganizationStore is a store that holds Organizations.
type OrganizationStore interface {
	// Create creates a new organization.
	Create(Organization) error

	// GetByID finds the organization by ID and retrieves it.
	GetByID(ttnpb.OrganizationIdentifiers, OrganizationSpecializer) (Organization, error)

	// ListByUser returns the organizations to which an user is a member of.
	ListByUser(ttnpb.UserIdentifiers, OrganizationSpecializer) ([]Organization, error)

	// Update updates an organization.
	Update(Organization) error

	// TODO(gomezjdaniel#274): use sql 'ON DELETE CASCADE' when CockroachDB implements it.
	// Delete deletes an organization.
	Delete(ttnpb.OrganizationIdentifiers) error

	// SaveAPIKey stores an API Key attached to an organization.
	SaveAPIKey(ttnpb.OrganizationIdentifiers, ttnpb.APIKey) error

	// GetAPIKey retrieves an API key by value and the organization identifiers.
	GetAPIKey(string) (ttnpb.OrganizationIdentifiers, ttnpb.APIKey, error)

	// GetAPIKeyByName retrieves an API key from an organization.
	GetAPIKeyByName(ttnpb.OrganizationIdentifiers, string) (ttnpb.APIKey, error)

	// UpdateAPIKeyRights updates the right of an API key.
	UpdateAPIKeyRights(ttnpb.OrganizationIdentifiers, string, []ttnpb.Right) error

	// ListAPIKeys list all the API keys that an organization has.
	ListAPIKeys(ttnpb.OrganizationIdentifiers) ([]ttnpb.APIKey, error)

	// DeleteAPIKey deletes a given API key from an organization.
	DeleteAPIKey(ttnpb.OrganizationIdentifiers, string) error

	// SetMember inserts or updates a member within an organization.
	// If the list of rights is empty the member will be unset.
	SetMember(ttnpb.OrganizationMember) error

	// HasMemberRights checks whether an user has or not a set of given rights to
	// an organization. Returns false if the user is not part of the organization.
	HasMemberRights(ttnpb.OrganizationIdentifiers, ttnpb.UserIdentifiers, ...ttnpb.Right) (bool, error)

	// ListMembers retrieves all the members from an organization. Optionally a
	// list of rights can be passed to filter them.
	ListMembers(ttnpb.OrganizationIdentifiers, ...ttnpb.Right) ([]ttnpb.OrganizationMember, error)

	// ListMemberRights returns the rights a given User has for an Organization.
	ListMemberRights(ttnpb.OrganizationIdentifiers, ttnpb.UserIdentifiers) ([]ttnpb.Right, error)

	// LoadAttributes loads extra attributes into the Organization.
	LoadAttributes(ttnpb.OrganizationIdentifiers, Organization) error

	// StoreAttributes writes the extra attributes on the Organization if it is an
	// Attributer to the store.
	StoreAttributes(ttnpb.OrganizationIdentifiers, Organization) error
}

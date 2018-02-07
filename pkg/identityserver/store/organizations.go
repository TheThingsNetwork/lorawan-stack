// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
	Create(org Organization) error

	// GetByID finds the organization by ID and retrieves it.
	GetByID(organizationID string, specializer OrganizationSpecializer) (Organization, error)

	// ListByUser returns the organizations to which an user is a member of.
	ListByUser(userID string, specializer OrganizationSpecializer) ([]Organization, error)

	// Update updates an organization.
	Update(org Organization) error

	// TODO(gomezjdaniel#274): use sql 'ON DELETE CASCADE' when CockroachDB implements it.
	// Delete deletes an organization.
	Delete(organizationID string) error

	// SaveAPIKey stores an API Key attached to an organization.
	SaveAPIKey(organizationID string, key *ttnpb.APIKey) error

	// GetAPIKey retrieves an API key by value and the organization ID.
	GetAPIKey(key string) (string, *ttnpb.APIKey, error)

	// GetAPIKeyByName retrieves an API key from an organization.
	GetAPIKeyByName(organizationID, keyName string) (*ttnpb.APIKey, error)

	// UpdateAPIKeyRights updates the right of an API key.
	UpdateAPIKeyRights(organizationID, keyName string, rights []ttnpb.Right) error

	// ListAPIKey list all the API keys that an organization has.
	ListAPIKeys(organizationID string) ([]*ttnpb.APIKey, error)

	// DeleteAPIKey deletes a given API key from an organization.
	DeleteAPIKey(organizationID, keyName string) error

	// SetMember inserts or updates a member within an organization.
	// If the list of rights is empty the member will be unset.
	SetMember(member *ttnpb.OrganizationMember) error

	// HasMemberRights checks whether an user has or not a set of given rights to
	// an organization. Returns false if the user is not part of the organization.
	HasMemberRights(organizationID, userID string, rights ...ttnpb.Right) (bool, error)

	// ListMembers retrieves all the members from an organization. Optionally a
	// list of rights can be passed to filter them.
	ListMembers(organizationID string, rights ...ttnpb.Right) ([]*ttnpb.OrganizationMember, error)

	// ListUserRights returns the rights a given User has for an Organization.
	ListUserRights(organizationID string, userID string) ([]ttnpb.Right, error)

	// LoadAttributes loads extra attributes into the Organization.
	LoadAttributes(organizationID string, org Organization) error

	// StoreAttributes writes the extra attributes on the Organization if it is an
	// Attributer to the store.
	StoreAttributes(organizationID string, org, result Organization) error
}

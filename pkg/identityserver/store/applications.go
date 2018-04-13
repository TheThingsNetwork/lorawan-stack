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

// Application is the interface of all things that can be an application. This can
// be used to build richer user types that can still be read and written to a database.
type Application interface {
	// GetApplication returns the ttnpb.Application that represents this
	// application.
	GetApplication() *ttnpb.Application
}

// ApplicationSpecializer returns a new Application with the given base ttnpb.Application.
type ApplicationSpecializer func(ttnpb.Application) Application

// ApplicationStore is a store that holds Applications.
// nolint: dupl
type ApplicationStore interface {
	// Create creates a new application.
	Create(Application) error

	// GetByID finds the application by ID and retrieves it.
	GetByID(ttnpb.ApplicationIdentifiers, ApplicationSpecializer) (Application, error)

	// ListByOrganizationOrUser returns the applications to which an organization
	// or user is collaborator of.
	ListByOrganizationOrUser(ttnpb.OrganizationOrUserIdentifiers, ApplicationSpecializer) ([]Application, error)

	// Update updates the application.
	Update(Application) error

	// Delete deletes an application.
	Delete(ttnpb.ApplicationIdentifiers) error

	// SaveAPIKey stores an API Key attached to an application.
	SaveAPIKey(ttnpb.ApplicationIdentifiers, ttnpb.APIKey) error

	// GetAPIKey retrieves an API key by value and the appplication identifiers.
	GetAPIKey(string) (ttnpb.ApplicationIdentifiers, ttnpb.APIKey, error)

	// GetAPIKeyByName retrieves an API key from an application.
	GetAPIKeyByName(ttnpb.ApplicationIdentifiers, string) (ttnpb.APIKey, error)

	// UpdateAPIKeyRights updates the right of an API key.
	// TODO(gomezjdaniel): merge with SaveAPIKey and rename it to `Set`?
	UpdateAPIKeyRights(ttnpb.ApplicationIdentifiers, string, []ttnpb.Right) error

	// ListAPIKeys list all the API keys that an application has.
	ListAPIKeys(ttnpb.ApplicationIdentifiers) ([]ttnpb.APIKey, error)

	// DeleteAPIKey deletes a given API key from an application.
	DeleteAPIKey(ttnpb.ApplicationIdentifiers, string) error

	// SetCollaborator inserts or updates a collaborator within an application.
	// If the list of rights is empty the collaborator will be unset.
	SetCollaborator(ttnpb.ApplicationCollaborator) error

	// HasCollaboratorRights checks whether a collaborator has a given set of rights
	// to an application. It returns false if the collaborationship does not exist.
	HasCollaboratorRights(ttnpb.ApplicationIdentifiers, ttnpb.OrganizationOrUserIdentifiers, ...ttnpb.Right) (bool, error)

	// ListCollaborators retrieves all the collaborators from an application.
	// Optionally a list of rights can be passed to filter them.
	ListCollaborators(ttnpb.ApplicationIdentifiers, ...ttnpb.Right) ([]ttnpb.ApplicationCollaborator, error)

	// ListCollaboratorRights returns the rights a given collaborator has for an
	// Application. Returns empty list if the collaborationship does not exist.
	ListCollaboratorRights(ttnpb.ApplicationIdentifiers, ttnpb.OrganizationOrUserIdentifiers) ([]ttnpb.Right, error)

	// LoadAttributes loads extra attributes into the Application.
	LoadAttributes(ttnpb.ApplicationIdentifiers, Application) error

	// StoreAttributes writes the extra attributes on the Application if it is an
	// Attributer to the store.
	StoreAttributes(ttnpb.ApplicationIdentifiers, Application) error
}

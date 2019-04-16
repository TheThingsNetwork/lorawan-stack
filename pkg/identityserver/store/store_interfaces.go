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

package store

import (
	"context"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// ApplicationStore interface for storing Applications.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type ApplicationStore interface {
	CreateApplication(ctx context.Context, app *ttnpb.Application) (*ttnpb.Application, error)
	FindApplications(ctx context.Context, ids []*ttnpb.ApplicationIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.Application, error)
	GetApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Application, error)
	UpdateApplication(ctx context.Context, app *ttnpb.Application, fieldMask *types.FieldMask) (*ttnpb.Application, error)
	DeleteApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) error
}

// ClientStore interface for storing Clients.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type ClientStore interface {
	CreateClient(ctx context.Context, cli *ttnpb.Client) (*ttnpb.Client, error)
	FindClients(ctx context.Context, ids []*ttnpb.ClientIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.Client, error)
	GetClient(ctx context.Context, id *ttnpb.ClientIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Client, error)
	UpdateClient(ctx context.Context, cli *ttnpb.Client, fieldMask *types.FieldMask) (*ttnpb.Client, error)
	DeleteClient(ctx context.Context, id *ttnpb.ClientIdentifiers) error
}

// EndDeviceStore interface for storing EndDevices.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type EndDeviceStore interface {
	CreateEndDevice(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, error)
	CountEndDevices(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (uint64, error)
	ListEndDevices(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.EndDevice, error)
	FindEndDevices(ctx context.Context, ids []*ttnpb.EndDeviceIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.EndDevice, error)
	GetEndDevice(ctx context.Context, id *ttnpb.EndDeviceIdentifiers, fieldMask *types.FieldMask) (*ttnpb.EndDevice, error)
	UpdateEndDevice(ctx context.Context, dev *ttnpb.EndDevice, fieldMask *types.FieldMask) (*ttnpb.EndDevice, error)
	DeleteEndDevice(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) error
}

// GatewayStore interface for storing Gateways.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type GatewayStore interface {
	CreateGateway(ctx context.Context, gtw *ttnpb.Gateway) (*ttnpb.Gateway, error)
	FindGateways(ctx context.Context, ids []*ttnpb.GatewayIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.Gateway, error)
	GetGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Gateway, error)
	UpdateGateway(ctx context.Context, gtw *ttnpb.Gateway, fieldMask *types.FieldMask) (*ttnpb.Gateway, error)
	DeleteGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error
}

// OrganizationStore interface for storing Organizations.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type OrganizationStore interface {
	CreateOrganization(ctx context.Context, org *ttnpb.Organization) (*ttnpb.Organization, error)
	FindOrganizations(ctx context.Context, ids []*ttnpb.OrganizationIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.Organization, error)
	GetOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Organization, error)
	UpdateOrganization(ctx context.Context, org *ttnpb.Organization, fieldMask *types.FieldMask) (*ttnpb.Organization, error)
	DeleteOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers) error
}

// UserStore interface for storing Users.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type UserStore interface {
	CreateUser(ctx context.Context, usr *ttnpb.User) (*ttnpb.User, error)
	FindUsers(ctx context.Context, ids []*ttnpb.UserIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.User, error)
	GetUser(ctx context.Context, id *ttnpb.UserIdentifiers, fieldMask *types.FieldMask) (*ttnpb.User, error)
	UpdateUser(ctx context.Context, usr *ttnpb.User, fieldMask *types.FieldMask) (*ttnpb.User, error)
	DeleteUser(ctx context.Context, id *ttnpb.UserIdentifiers) error
}

// UserSessionStore interface for storing User sessions.
//
// For internal use (by the OAuth server) only.
type UserSessionStore interface {
	CreateSession(ctx context.Context, sess *ttnpb.UserSession) (*ttnpb.UserSession, error)
	FindSessions(ctx context.Context, userIDs *ttnpb.UserIdentifiers) ([]*ttnpb.UserSession, error)
	GetSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) (*ttnpb.UserSession, error)
	UpdateSession(ctx context.Context, sess *ttnpb.UserSession) (*ttnpb.UserSession, error)
	DeleteSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) error
}

// MembershipStore interface for storing membership (collaboration) relations
// between accounts (users or organizations) and entities (applications, clients,
// gateways or organizations).
//
// As the operations in this store may be quite expensive, the results of FindXXX
// operations should typically be cached. The recommended cache behavior is:
//
// - The results of FindMembers are cached by (entity_type,entity_id)
// - The results of FindMemberRights are cached by (account_id,[entityType])
// - Any (successful or unsuccessful) call to SetMember expires the caches
//   for (entity_type,entity_id) and (account_id,[entityType]).
type MembershipStore interface {
	// Find direct members and rights of the given entity.
	FindMembers(ctx context.Context, entityID *ttnpb.EntityIdentifiers) (map[*ttnpb.OrganizationOrUserIdentifiers]*ttnpb.Rights, error)
	// Find direct member rights of the given organization or user. The entityType may be omitted.
	FindMemberRights(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityType string) (map[*ttnpb.EntityIdentifiers]*ttnpb.Rights, error)
	// Set member rights on an entity. Rights can be deleted by not passing any rights.
	SetMember(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityID *ttnpb.EntityIdentifiers, rights *ttnpb.Rights) error
}

// APIKeyStore interface for storing API keys for entities (applications,
// clients, gateways, organizations or users).
type APIKeyStore interface {
	// Create a new API key for the given entity.
	CreateAPIKey(ctx context.Context, entityID *ttnpb.EntityIdentifiers, key *ttnpb.APIKey) error
	// Find API keys of the given entity.
	FindAPIKeys(ctx context.Context, entityID *ttnpb.EntityIdentifiers) ([]*ttnpb.APIKey, error)
	// Get an API key by its ID.
	GetAPIKey(ctx context.Context, id string) (*ttnpb.EntityIdentifiers, *ttnpb.APIKey, error)
	// Update key rights on an entity. Rights can be deleted by not passing any rights, in which case the returned API key will be nil.
	UpdateAPIKey(ctx context.Context, entityID *ttnpb.EntityIdentifiers, key *ttnpb.APIKey) (*ttnpb.APIKey, error)
}

// OAuthStore interface for the OAuth server.
//
// For internal use (by the OAuth server) only.
type OAuthStore interface {
	ListAuthorizations(ctx context.Context, userIDs *ttnpb.UserIdentifiers) ([]*ttnpb.OAuthClientAuthorization, error)
	GetAuthorization(ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers) (*ttnpb.OAuthClientAuthorization, error)
	Authorize(ctx context.Context, req *ttnpb.OAuthClientAuthorization) (authorization *ttnpb.OAuthClientAuthorization, err error)
	DeleteAuthorization(ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers) error

	CreateAuthorizationCode(ctx context.Context, code *ttnpb.OAuthAuthorizationCode) error
	GetAuthorizationCode(ctx context.Context, code string) (*ttnpb.OAuthAuthorizationCode, error)
	DeleteAuthorizationCode(ctx context.Context, code string) error

	CreateAccessToken(ctx context.Context, token *ttnpb.OAuthAccessToken, previousID string) error
	ListAccessTokens(ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers) ([]*ttnpb.OAuthAccessToken, error)
	GetAccessToken(ctx context.Context, id string) (*ttnpb.OAuthAccessToken, error)
	DeleteAccessToken(ctx context.Context, id string) error
}

// InvitationStore interface for storing user invitations.
type InvitationStore interface {
	CreateInvitation(ctx context.Context, invitation *ttnpb.Invitation) (*ttnpb.Invitation, error)
	FindInvitations(ctx context.Context) ([]*ttnpb.Invitation, error)
	GetInvitation(ctx context.Context, token string) (*ttnpb.Invitation, error)
	SetInvitationAcceptedBy(ctx context.Context, token string, accepter *ttnpb.UserIdentifiers) error
	DeleteInvitation(ctx context.Context, email string) error
}

// EntitySearch interface for searching entities.
type EntitySearch interface {
	FindEntities(ctx context.Context, req *ttnpb.SearchEntitiesRequest, entityType string) ([]*ttnpb.EntityIdentifiers, error)
}

// ContactInfoStore interface for contact info validation.
type ContactInfoStore interface {
	GetContactInfo(ctx context.Context, entityID *ttnpb.EntityIdentifiers) ([]*ttnpb.ContactInfo, error)
	SetContactInfo(ctx context.Context, entityID *ttnpb.EntityIdentifiers, contactInfo []*ttnpb.ContactInfo) ([]*ttnpb.ContactInfo, error)
	CreateValidation(ctx context.Context, validation *ttnpb.ContactInfoValidation) (*ttnpb.ContactInfoValidation, error)
	// Confirm a validation. Only the ID and Token need to be set.
	Validate(ctx context.Context, validation *ttnpb.ContactInfoValidation) error
}

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

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// ApplicationStore interface for storing Applications.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type ApplicationStore interface {
	CreateApplication(ctx context.Context, app *ttnpb.Application) (*ttnpb.Application, error)
	CountApplications(ctx context.Context) (uint64, error)
	FindApplications(
		ctx context.Context, ids []*ttnpb.ApplicationIdentifiers, fieldMask FieldMask,
	) ([]*ttnpb.Application, error)
	GetApplication(
		ctx context.Context, id *ttnpb.ApplicationIdentifiers, fieldMask FieldMask,
	) (*ttnpb.Application, error)
	UpdateApplication(
		ctx context.Context, app *ttnpb.Application, fieldMask FieldMask,
	) (*ttnpb.Application, error)
	DeleteApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) error
	RestoreApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) error
	PurgeApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) error
}

// ClientStore interface for storing Clients.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type ClientStore interface {
	CreateClient(ctx context.Context, cli *ttnpb.Client) (*ttnpb.Client, error)
	CountClients(ctx context.Context) (uint64, error)
	FindClients(
		ctx context.Context, ids []*ttnpb.ClientIdentifiers, fieldMask FieldMask,
	) ([]*ttnpb.Client, error)
	GetClient(
		ctx context.Context, id *ttnpb.ClientIdentifiers, fieldMask FieldMask,
	) (*ttnpb.Client, error)
	UpdateClient(
		ctx context.Context, cli *ttnpb.Client, fieldMask FieldMask,
	) (*ttnpb.Client, error)
	DeleteClient(ctx context.Context, id *ttnpb.ClientIdentifiers) error
	RestoreClient(ctx context.Context, id *ttnpb.ClientIdentifiers) error
	PurgeClient(ctx context.Context, id *ttnpb.ClientIdentifiers) error
}

// EndDeviceStore interface for storing EndDevices.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type EndDeviceStore interface {
	CreateEndDevice(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, error)
	CountEndDevices(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (uint64, error)
	ListEndDevices(
		ctx context.Context, ids *ttnpb.ApplicationIdentifiers, fieldMask FieldMask,
	) ([]*ttnpb.EndDevice, error)
	FindEndDevices(
		ctx context.Context, ids []*ttnpb.EndDeviceIdentifiers, fieldMask FieldMask,
	) ([]*ttnpb.EndDevice, error)
	GetEndDevice(
		ctx context.Context, id *ttnpb.EndDeviceIdentifiers, fieldMask FieldMask,
	) (*ttnpb.EndDevice, error)
	UpdateEndDevice(
		ctx context.Context, dev *ttnpb.EndDevice, fieldMask FieldMask,
	) (*ttnpb.EndDevice, error)
	DeleteEndDevice(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) error
	BatchUpdateEndDeviceLastSeen(
		ctx context.Context,
		devsLastSeen []*ttnpb.BatchUpdateEndDeviceLastSeenRequest_EndDeviceLastSeenUpdate,
	) error
}

// GatewayStore interface for storing Gateways.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type GatewayStore interface {
	CreateGateway(ctx context.Context, gtw *ttnpb.Gateway) (*ttnpb.Gateway, error)
	CountGateways(ctx context.Context) (uint64, error)
	FindGateways(
		ctx context.Context, ids []*ttnpb.GatewayIdentifiers, fieldMask FieldMask,
	) ([]*ttnpb.Gateway, error)
	GetGateway(
		ctx context.Context, id *ttnpb.GatewayIdentifiers, fieldMask FieldMask,
	) (*ttnpb.Gateway, error)
	UpdateGateway(
		ctx context.Context, gtw *ttnpb.Gateway, fieldMask FieldMask,
	) (*ttnpb.Gateway, error)
	DeleteGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error
	RestoreGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error
	PurgeGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error
}

// OrganizationStore interface for storing Organizations.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type OrganizationStore interface {
	CreateOrganization(
		ctx context.Context, org *ttnpb.Organization,
	) (*ttnpb.Organization, error)
	CountOrganizations(ctx context.Context) (uint64, error)
	FindOrganizations(
		ctx context.Context, ids []*ttnpb.OrganizationIdentifiers, fieldMask FieldMask,
	) ([]*ttnpb.Organization, error)
	GetOrganization(
		ctx context.Context, id *ttnpb.OrganizationIdentifiers, fieldMask FieldMask,
	) (*ttnpb.Organization, error)
	UpdateOrganization(
		ctx context.Context, org *ttnpb.Organization, fieldMask FieldMask,
	) (*ttnpb.Organization, error)
	DeleteOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers) error
	RestoreOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers) error
	PurgeOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers) error
}

// UserStore interface for storing Users.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type UserStore interface {
	CreateUser(ctx context.Context, usr *ttnpb.User) (*ttnpb.User, error)
	CountUsers(ctx context.Context) (uint64, error)
	FindUsers(
		ctx context.Context, ids []*ttnpb.UserIdentifiers, fieldMask FieldMask,
	) ([]*ttnpb.User, error)
	ListAdmins(ctx context.Context, fieldMask FieldMask) ([]*ttnpb.User, error)
	GetUser(
		ctx context.Context, id *ttnpb.UserIdentifiers, fieldMask FieldMask,
	) (*ttnpb.User, error)
	GetUserByPrimaryEmailAddress(
		ctx context.Context, email string, fieldMask FieldMask,
	) (*ttnpb.User, error)
	UpdateUser(
		ctx context.Context, usr *ttnpb.User, fieldMask FieldMask,
	) (*ttnpb.User, error)
	DeleteUser(ctx context.Context, id *ttnpb.UserIdentifiers) error
	RestoreUser(ctx context.Context, id *ttnpb.UserIdentifiers) error
	PurgeUser(ctx context.Context, id *ttnpb.UserIdentifiers) error
}

// UserSessionStore interface for storing User sessions.
//
// For internal use (by the OAuth server) only.
type UserSessionStore interface {
	CreateSession(
		ctx context.Context, sess *ttnpb.UserSession,
	) (*ttnpb.UserSession, error)
	FindSessions(
		ctx context.Context, userIDs *ttnpb.UserIdentifiers,
	) ([]*ttnpb.UserSession, error)
	GetSession(
		ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string,
	) (*ttnpb.UserSession, error)
	GetSessionByID(ctx context.Context, tokenID string) (*ttnpb.UserSession, error)
	DeleteSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) error
	DeleteAllUserSessions(ctx context.Context, userIDs *ttnpb.UserIdentifiers) error
}

// MembershipStore interface for storing membership (collaboration) relations
// between accounts (users or organizations) and entities (applications, clients,
// gateways or organizations).
//
// As the operations in this store may be quite expensive, the results of FindXXX
// operations should typically be cached. The recommended cache behavior is:
type MembershipStore interface {
	// Count direct memberships of the organization or user.
	CountMemberships(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityType string) (uint64, error)

	// Find direct and optionally also indirect memberships of the organization or user.
	FindMemberships(
		ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityType string, includeIndirect bool,
	) ([]*ttnpb.EntityIdentifiers, error)

	// Find memberships (through organizations) between the user and entity.
	FindAccountMembershipChains(
		ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, entityType string, entityIDs ...string,
	) ([]*MembershipChain, error)

	// Find direct members and rights of the given entity.
	FindMembers(
		ctx context.Context, entityID *ttnpb.EntityIdentifiers,
	) ([]*MemberByID, error)
	// Get direct member rights on an entity.
	GetMember(
		ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityID *ttnpb.EntityIdentifiers,
	) (*ttnpb.Rights, error)
	// Set direct member rights on an entity.
	SetMember(
		ctx context.Context,
		id *ttnpb.OrganizationOrUserIdentifiers,
		entityID *ttnpb.EntityIdentifiers,
		rights *ttnpb.Rights,
	) error
	// DeleteMember elminates the direct member rights attached to an entity.
	DeleteMember(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityID *ttnpb.EntityIdentifiers) error
	// Delete all member rights on an entity. Used for purging entities.
	DeleteEntityMembers(ctx context.Context, entityID *ttnpb.EntityIdentifiers) error
	// Delete all user rights for an entity.
	DeleteAccountMembers(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers) error
}

// APIKeyStore interface for storing API keys for entities (applications,
// clients, gateways, organizations or users).
type APIKeyStore interface {
	// Create a new API key for the given entity.
	CreateAPIKey(
		ctx context.Context, entityID *ttnpb.EntityIdentifiers, key *ttnpb.APIKey,
	) (*ttnpb.APIKey, error)
	// Find API keys of the given entity.
	FindAPIKeys(
		ctx context.Context, entityID *ttnpb.EntityIdentifiers,
	) ([]*ttnpb.APIKey, error)
	// Get an API key.
	GetAPIKey(
		ctx context.Context, entityID *ttnpb.EntityIdentifiers, id string,
	) (*ttnpb.APIKey, error)
	// Get an API key by its ID.
	GetAPIKeyByID(
		ctx context.Context, id string,
	) (*ttnpb.EntityIdentifiers, *ttnpb.APIKey, error)
	// Update key rights on an entity.
	// Rights can be deleted by not passing any rights, in which case the returned API key will be nil.
	UpdateAPIKey(
		ctx context.Context, entityID *ttnpb.EntityIdentifiers, key *ttnpb.APIKey, fieldMask FieldMask,
	) (*ttnpb.APIKey, error)
	// DeleteAPIKey deletes key rights on an entity.
	DeleteAPIKey(ctx context.Context, entityID *ttnpb.EntityIdentifiers, key *ttnpb.APIKey) error
	// Delete api keys deletes all api keys tied to an entity. Used when purging entities.
	DeleteEntityAPIKeys(ctx context.Context, entityID *ttnpb.EntityIdentifiers) error
}

// OAuthStore interface for the OAuth server.
//
// For internal use (by the OAuth server) only.
type OAuthStore interface {
	ListAuthorizations(
		ctx context.Context, userIDs *ttnpb.UserIdentifiers,
	) ([]*ttnpb.OAuthClientAuthorization, error)
	GetAuthorization(
		ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
	) (*ttnpb.OAuthClientAuthorization, error)
	Authorize(
		ctx context.Context, req *ttnpb.OAuthClientAuthorization,
	) (authorization *ttnpb.OAuthClientAuthorization, err error)
	DeleteAuthorization(
		ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
	) error
	DeleteUserAuthorizations(ctx context.Context, userIDs *ttnpb.UserIdentifiers) error
	DeleteClientAuthorizations(ctx context.Context, clientIDs *ttnpb.ClientIdentifiers) error

	CreateAuthorizationCode(
		ctx context.Context, code *ttnpb.OAuthAuthorizationCode,
	) (*ttnpb.OAuthAuthorizationCode, error)
	GetAuthorizationCode(ctx context.Context, code string) (*ttnpb.OAuthAuthorizationCode, error)
	DeleteAuthorizationCode(ctx context.Context, code string) error

	CreateAccessToken(
		ctx context.Context, token *ttnpb.OAuthAccessToken, previousID string,
	) (*ttnpb.OAuthAccessToken, error)
	ListAccessTokens(
		ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
	) ([]*ttnpb.OAuthAccessToken, error)
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

// LoginTokenStore interface for storing user login tokens.
type LoginTokenStore interface {
	FindActiveLoginTokens(ctx context.Context, userIDs *ttnpb.UserIdentifiers) ([]*ttnpb.LoginToken, error)
	CreateLoginToken(ctx context.Context, token *ttnpb.LoginToken) (*ttnpb.LoginToken, error)
	ConsumeLoginToken(ctx context.Context, token string) (*ttnpb.LoginToken, error)
}

// EntitySearch interface for searching entities.
type EntitySearch interface {
	SearchApplications(
		ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchApplicationsRequest,
	) ([]*ttnpb.ApplicationIdentifiers, error)
	SearchClients(
		ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchClientsRequest,
	) ([]*ttnpb.ClientIdentifiers, error)
	SearchEndDevices(
		ctx context.Context, req *ttnpb.SearchEndDevicesRequest,
	) ([]*ttnpb.EndDeviceIdentifiers, error)
	SearchGateways(
		ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchGatewaysRequest,
	) ([]*ttnpb.GatewayIdentifiers, error)
	SearchOrganizations(
		ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchOrganizationsRequest,
	) ([]*ttnpb.OrganizationIdentifiers, error)
	SearchUsers(
		ctx context.Context, req *ttnpb.SearchUsersRequest,
	) ([]*ttnpb.UserIdentifiers, error)
	SearchAccounts(
		ctx context.Context, req *ttnpb.SearchAccountsRequest,
	) ([]*ttnpb.OrganizationOrUserIdentifiers, error)
}

// ContactInfoStore interface for contact info validation.
type ContactInfoStore interface {
	GetContactInfo(ctx context.Context, entityID ttnpb.IDStringer) ([]*ttnpb.ContactInfo, error)
	SetContactInfo(
		ctx context.Context, entityID ttnpb.IDStringer, contactInfo []*ttnpb.ContactInfo,
	) ([]*ttnpb.ContactInfo, error)
	ValidateContactInfo(ctx context.Context, validation *ttnpb.ContactInfoValidation) error
	DeleteEntityContactInfo(ctx context.Context, entityID ttnpb.IDStringer) error

	CreateValidation(ctx context.Context, validation *ttnpb.ContactInfoValidation) (*ttnpb.ContactInfoValidation, error)
	GetValidation(ctx context.Context, validation *ttnpb.ContactInfoValidation) (*ttnpb.ContactInfoValidation, error)
	ExpireValidation(ctx context.Context, validation *ttnpb.ContactInfoValidation) error
}

// EUIStore interface for assigning DevEUI blocks and addresses.
type EUIStore interface {
	CreateEUIBlock(
		ctx context.Context, configPrefix types.EUI64Prefix, initCounter int64, euiType string,
	) error
	IssueDevEUIForApplication(
		ctx context.Context, id *ttnpb.ApplicationIdentifiers, applicationLimit int,
	) (*types.EUI64, error)
}

// NotificationStore interface for notifications.
type NotificationStore interface {
	CreateNotification(
		ctx context.Context, notification *ttnpb.Notification, receiverIDs []*ttnpb.UserIdentifiers,
	) (*ttnpb.Notification, error)
	ListNotifications(
		ctx context.Context, receiverIDs *ttnpb.UserIdentifiers, statuses []ttnpb.NotificationStatus,
	) ([]*ttnpb.Notification, error)
	UpdateNotificationStatus(
		ctx context.Context,
		receiverIDs *ttnpb.UserIdentifiers,
		notificationIDs []string,
		status ttnpb.NotificationStatus,
	) error
}

// Store interface combines the interfaces of all individual stores.
type Store interface {
	ApplicationStore
	ClientStore
	EndDeviceStore
	GatewayStore
	OrganizationStore
	UserStore
	UserSessionStore
	MembershipStore
	APIKeyStore
	OAuthStore
	InvitationStore
	LoginTokenStore
	ContactInfoStore
	EUIStore
	NotificationStore
	EntitySearch
}

// TransactionalStore is Store, but with a method that uses a transaction.
type TransactionalStore interface {
	Store

	Transact(ctx context.Context, fc func(context.Context, Store) error) error
}

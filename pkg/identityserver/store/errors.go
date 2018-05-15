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

import "go.thethings.network/lorawan-stack/pkg/errors"

func init() {
	ErrApplicationNotFound.Register()
	ErrApplicationIDTaken.Register()

	ErrClientNotFound.Register()
	ErrClientIDTaken.Register()

	ErrGatewayNotFound.Register()
	ErrGatewayIDTaken.Register()

	ErrAuthorizationCodeNotFound.Register()
	ErrAuthorizationCodeConflict.Register()
	ErrRefreshTokenNotFound.Register()
	ErrRefreshTokenConflict.Register()
	ErrAccessTokenNotFound.Register()
	ErrAccessTokenConflict.Register()
	ErrAuthorizedClientNotFound.Register()

	ErrUserNotFound.Register()
	ErrUserIDTaken.Register()
	ErrUserEmailTaken.Register()
	ErrValidationTokenNotFound.Register()

	ErrAPIKeyNotFound.Register()
	ErrAPIKeyNameConflict.Register()

	ErrSettingsNotFound.Register()

	ErrInvitationNotFound.Register()

	ErrOrganizationIDTaken.Register()
	ErrOrganizationNotFound.Register()
}

// ErrApplicationNotFound is returned when trying to fetch an application that
// does not exist.
var ErrApplicationNotFound = &errors.ErrDescriptor{
	MessageFormat: "Application does not exist",
	Code:          1,
	Type:          errors.NotFound,
}

// ErrApplicationIDTaken is returned when trying to create a new application
// with an ID that already exists.
var ErrApplicationIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Application id is already taken",
	Code:          2,
	Type:          errors.AlreadyExists,
}

// ErrClientNotFound is returned when trying to fetch a client that does not exists.
var ErrClientNotFound = &errors.ErrDescriptor{
	MessageFormat: "Client does not exist",
	Code:          3,
	Type:          errors.NotFound,
}

// ErrClientIDTaken is returned when trying to create a new client with an ID
// that already exists
var ErrClientIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Client id is already taken",
	Code:          4,
	Type:          errors.AlreadyExists,
}

// ErrGatewayNotFound is returned when trying to fetch a gateway that does not exist.
var ErrGatewayNotFound = &errors.ErrDescriptor{
	MessageFormat: "Gateway does not exist",
	Code:          5,
	Type:          errors.NotFound,
}

// ErrGatewayIDTaken is returned when trying to create a new gateway with an ID
// that already exists.
var ErrGatewayIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Gateway id is already taken",
	Code:          6,
	Type:          errors.AlreadyExists,
}

// ErrAuthorizationCodeNotFound is returned when trying to fetch or delete an
// authorization code that does not exist.
var ErrAuthorizationCodeNotFound = &errors.ErrDescriptor{
	MessageFormat: "Authorization code does not exist",
	Code:          7,
	Type:          errors.NotFound,
}

// ErrAuthorizationCodeConflict is returned when trying to create an authorization
// code that already exists.
var ErrAuthorizationCodeConflict = &errors.ErrDescriptor{
	MessageFormat: "Authorization code already exists",
	Code:          8,
	Type:          errors.AlreadyExists,
}

// ErrRefreshTokenNotFound is returned when trying to fetch or delete a refresh token
// that does not exist.
var ErrRefreshTokenNotFound = &errors.ErrDescriptor{
	MessageFormat: "Refresh token does not exist",
	Code:          9,
	Type:          errors.NotFound,
}

// ErrRefreshTokenConflict is returned when trying to create a refresh token
// that already exists.
var ErrRefreshTokenConflict = &errors.ErrDescriptor{
	MessageFormat: "Refresh token already exists",
	Code:          10,
	Type:          errors.AlreadyExists,
}

// ErrAccessTokenNotFound is returned when trying to fetch or delete an access
// token that does not exist.
var ErrAccessTokenNotFound = &errors.ErrDescriptor{
	MessageFormat: "Access token does not exist",
	Code:          11,
	Type:          errors.NotFound,
}

// ErrAccessTokenConflict is returned when trying to create an access token that
// already exists.
var ErrAccessTokenConflict = &errors.ErrDescriptor{
	MessageFormat: "Access token already exists",
	Code:          12,
	Type:          errors.AlreadyExists,
}

// ErrAuthorizedClientNotFound is returned when an user tries to revoke a
// non-existent authorized client.
var ErrAuthorizedClientNotFound = &errors.ErrDescriptor{
	MessageFormat: "Client is not authorized",
	Code:          13,
	Type:          errors.NotFound,
}

// ErrUserNotFound is returned when trying to fetch an user that does not exist.
var ErrUserNotFound = &errors.ErrDescriptor{
	MessageFormat: "User does not exist",
	Code:          14,
	Type:          errors.NotFound,
}

// ErrUserIDTaken is returned when trying to create a new user with an ID that
// is already taken.
var ErrUserIDTaken = &errors.ErrDescriptor{
	MessageFormat: "User id is already taken",
	Code:          15,
	Type:          errors.AlreadyExists,
}

// ErrUserEmailTaken is returned when trying to create a new user with an
// email that is already taken.
var ErrUserEmailTaken = &errors.ErrDescriptor{
	MessageFormat: "Email address is already taken by another account",
	Code:          16,
	Type:          errors.Conflict,
}

// ErrValidationTokenNotFound is returned when trying to get a token that does not exist.
var ErrValidationTokenNotFound = &errors.ErrDescriptor{
	MessageFormat: "Validation token not found",
	Code:          17,
	Type:          errors.NotFound,
}

// ErrAPIKeyNotFound is returned when trying to fetch an API key that does not exist.
var ErrAPIKeyNotFound = &errors.ErrDescriptor{
	MessageFormat: "API key does not exist",
	Code:          18,
	Type:          errors.NotFound,
}

// ErrAPIKeyNameConflict is returned when trying to create an API key with a name
// that already exists in the namespace.
var ErrAPIKeyNameConflict = &errors.ErrDescriptor{
	MessageFormat: "API key name is already taken",
	Code:          19,
	Type:          errors.Conflict,
}

// ErrSettingsNotFound is returned when trying to fetch the Identity Server
// settings from the database but these are not set.
var ErrSettingsNotFound = &errors.ErrDescriptor{
	MessageFormat: "Settings not found",
	Code:          20,
	Type:          errors.NotFound,
}

// ErrInvitationNotFound is returned when trying to use an invitation that does not exist.
var ErrInvitationNotFound = &errors.ErrDescriptor{
	MessageFormat: "Invitation not found",
	Code:          21,
	Type:          errors.NotFound,
}

// ErrOrganizationIDTaken is returned when trying to create an organization
// using an ID that is already in use by an user or organization.
var ErrOrganizationIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Organization id is already taken",
	Code:          22,
	Type:          errors.AlreadyExists,
}

// ErrOrganizationNotFound is returned when trying to fetch an organization that
// does not exists.
var ErrOrganizationNotFound = &errors.ErrDescriptor{
	MessageFormat: "Organization does not exist",
	Code:          23,
	Type:          errors.NotFound,
}

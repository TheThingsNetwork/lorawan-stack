// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrApplicationNotFound.Register()
	ErrApplicationIDTaken.Register()
	ErrApplicationAPIKeyNotFound.Register()

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

	ErrUserNotFound.Register()
	ErrUserEmailNotFound.Register()
	ErrUserIDTaken.Register()
	ErrUserEmailTaken.Register()
}

// ErrApplicationNotFound is returned when trying to fetch an application that
// does not exist.
var ErrApplicationNotFound = &errors.ErrDescriptor{
	MessageFormat: "Application `{application_id}` does not exist",
	Code:          1,
	Type:          errors.NotFound,
	SafeAttributes: []string{
		"application_id",
	},
}

// ErrApplicationIDTaken is returned when trying to create a new application
// with an ID that already exists.
var ErrApplicationIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Application id `{application_id}` is already taken",
	Code:          2,
	Type:          errors.AlreadyExists,
	SafeAttributes: []string{
		"application_id",
	},
}

// ErrApplicationAPIKeyNotFound is returned when trying to access or delete
// an application API key that does not exist.
var ErrApplicationAPIKeyNotFound = &errors.ErrDescriptor{
	MessageFormat: "API key `{key_name}` does not exist for application `{application_id}`",
	Code:          3,
	Type:          errors.NotFound,
	SafeAttributes: []string{
		"application_id",
	},
}

// ErrClientNotFound is returned when trying to fetch a client that does not exists.
var ErrClientNotFound = &errors.ErrDescriptor{
	MessageFormat: "Client `{client_id}` does not exist",
	Code:          20,
	Type:          errors.NotFound,
	SafeAttributes: []string{
		"client_id",
	},
}

// ErrClientIDTaken is returned when trying to create a new client with an ID.
// that already exists
var ErrClientIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Client id `{client_id}` is already taken",
	Code:          21,
	Type:          errors.AlreadyExists,
	SafeAttributes: []string{
		"client_id",
	},
}

// ErrGatewayNotFound is returned when trying to fetch a gateway that does not exist.
var ErrGatewayNotFound = &errors.ErrDescriptor{
	MessageFormat: "Gateway `{gateway_id}` does not exist",
	Code:          300,
	Type:          errors.NotFound,
	SafeAttributes: []string{
		"gateway_id",
	},
}

// ErrGatewayIDTaken is returned when trying to create a new gateway with an ID
// that already exists.
var ErrGatewayIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Gateway id `{gateway_id}` is already taken",
	Code:          301,
	Type:          errors.AlreadyExists,
	SafeAttributes: []string{
		"gateway_id",
	},
}

// ErrAuthorizationCodeNotFound is returned when trying to fetch or delete an authorization code that
// does not exist.
var ErrAuthorizationCodeNotFound = &errors.ErrDescriptor{
	MessageFormat: "Authorization code token does not exist",
	Code:          501,
	Type:          errors.NotFound,
}

// ErrAuthorizationCodeConflict is returned when trying to create an authorization code that already exists.
var ErrAuthorizationCodeConflict = &errors.ErrDescriptor{
	MessageFormat: "Authorization code already exists",
	Code:          502,
	Type:          errors.Conflict,
}

// ErrRefreshTokenNotFound is returned when trying to fetch or delete a refresh token that
// does not exist.
var ErrRefreshTokenNotFound = &errors.ErrDescriptor{
	MessageFormat: "Refresh token does not exist",
	Code:          503,
	Type:          errors.NotFound,
}

// ErrRefreshTokenConflict is returned when trying to create a refresh token that already exists.
var ErrRefreshTokenConflict = &errors.ErrDescriptor{
	MessageFormat: "Refresh token already exists",
	Code:          504,
	Type:          errors.Conflict,
}

// ErrAccessTokenNotFound is returned when trying to fetch or delete an access
// token that does not exist.
var ErrAccessTokenNotFound = &errors.ErrDescriptor{
	MessageFormat: "Access token does not exist",
	Code:          505,
	Type:          errors.NotFound,
}

// ErrAccessTokenConflict is returned when trying to create an access token that
// already exists.
var ErrAccessTokenConflict = &errors.ErrDescriptor{
	MessageFormat: "Access token already exists",
	Code:          506,
	Type:          errors.Conflict,
}

// ErrUserNotFound is returned when trying to fetch an user that does not exist.
var ErrUserNotFound = &errors.ErrDescriptor{
	MessageFormat: "User `{user_id}` does not exist",
	Code:          400,
	Type:          errors.NotFound,
	SafeAttributes: []string{
		"user_id",
	},
}

// ErrUserEmailNotFound is returned when trying to find an user with an email
// that does not exist.
var ErrUserEmailNotFound = &errors.ErrDescriptor{
	MessageFormat: "User with email address `{email}` does not exist",
	Code:          401,
	Type:          errors.NotFound,
	SafeAttributes: []string{
		"email",
	},
}

// ErrUserIDTaken is returned when trying to create a new user with an ID that
// is already taken.
var ErrUserIDTaken = &errors.ErrDescriptor{
	MessageFormat: "User ID `{user_id}` is already taken",
	Code:          402,
	Type:          errors.AlreadyExists,
	SafeAttributes: []string{
		"user_id",
	},
}

// ErrUserEmailTaken is returned when trying to create a new user with an
// email that is already taken.
var ErrUserEmailTaken = &errors.ErrDescriptor{
	MessageFormat: "Email address `{email}` is already taken by another account",
	Code:          403,
	Type:          errors.AlreadyExists,
	SafeAttributes: []string{
		"email",
	},
}

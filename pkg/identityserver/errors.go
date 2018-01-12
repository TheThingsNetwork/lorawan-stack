// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrBlacklistedID.Register()
	ErrEmailAddressNotAllowed.Register()
	ErrInvalidPassword.Register()
	ErrNotAuthorized.Register()
	ErrSetApplicationCollaboratorFailed.Register()
	ErrSetGatewayCollaboratorFailed.Register()
	ErrRemoveGatewayAPIKeyFailed.Register()
	ErrEmailAlreadyValidated.Register()
	ErrValidationTokenExpired.Register()
}

// ErrBlacklistedID is returned when trying to register an entity using a blacklisted ID.
var ErrBlacklistedID = &errors.ErrDescriptor{
	MessageFormat:  "`{id} is not an allowed ID",
	Code:           1,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"id"},
}

// ErrEmailAddressNotAllowed is returned when trying to set in a user account an
// email address that is not allowed
var ErrEmailAddressNotAllowed = &errors.ErrDescriptor{
	MessageFormat:  "Email address `{email}` is not allowed. Current allowed email addresses by the network administrators are in the following form: {allowed_emails}",
	Code:           2,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"email", "allowed_emails"},
}

// ErrInvalidPassword is returned when trying to update the user's account
// password while providing a wrong current password.
var ErrInvalidPassword = &errors.ErrDescriptor{
	MessageFormat: "Invalid password",
	Code:          3,
	Type:          errors.Unauthorized,
}

// ErrNotAuthorized is returned when trying to access a protected resource without
// authorization.
var ErrNotAuthorized = &errors.ErrDescriptor{
	MessageFormat: "Not authorized",
	Code:          4,
	Type:          errors.Unauthorized,
}

// ErrSetApplicationCollaboratorFailed is returned when after modifying the
// collaborators of an application there is no collaborator with
// `RIGHT_APPLICATION_SETTINGS_COLLABORATORS` right.
var ErrSetApplicationCollaboratorFailed = &errors.ErrDescriptor{
	MessageFormat:  "Failed to modify collaborators: application `{application_id}` must have at least one collaborator with `RIGHT_APPLICATION_SETTINGS_COLLABORATORS` right",
	Code:           5,
	Type:           errors.Conflict,
	SafeAttributes: []string{"application_id"},
}

// ErrSetGatewayCollaboratorFailed is returned when after modifying the
// collaborators of a gateway there is no collaborator with
// `RIGHT_GATEWAY_SETTINGS_COLLABORATORS` right.
var ErrSetGatewayCollaboratorFailed = &errors.ErrDescriptor{
	MessageFormat:  "Failed to modify collaborators: gateway `{gateway_id}` must have at least one collaborator with `RIGHT_GATEWAY_SETTINGS_COLLABORATORS` right",
	Code:           6,
	Type:           errors.Conflict,
	SafeAttributes: []string{"gateway_id"},
}

// ErrRemoveGatewayAPIKeyFailed is returned when trying to remove the API key
// that's used as value to be sent to the gateway agent.
var ErrRemoveGatewayAPIKeyFailed = &errors.ErrDescriptor{
	MessageFormat:  "Failed to remove gateway `{gateway_id}` API key `{key_name}`: key must be rotated",
	Code:           7,
	Type:           errors.Unauthorized,
	SafeAttributes: []string{"gateway_id", "key_name"},
}

// ErrEmailAlreadyValidated is returned when calling 'RequestUserEmailValidation'
// when the email is already validated.
var ErrEmailAlreadyValidated = &errors.ErrDescriptor{
	MessageFormat:  "Email `{email}` is already validated",
	Code:           8,
	Type:           errors.Unauthorized,
	SafeAttributes: []string{"email"},
}

// ErrValidationTokenExpired is returned when trying to validate an email address
// using an expired token.
var ErrValidationTokenExpired = &errors.ErrDescriptor{
	MessageFormat: "Token is expired",
	Code:          9,
	Type:          errors.Unauthorized,
}

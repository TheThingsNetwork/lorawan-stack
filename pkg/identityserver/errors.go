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
	ErrEmailAlreadyValidated.Register()
	ErrValidationTokenExpired.Register()
	ErrInvitationTokenMissing.Register()
	ErrEmailAddressAlreadyUsed.Register()
	ErrSetOrganizationMemberFailed.Register()
}

// ErrBlacklistedID is returned when trying to register an entity using a
// blacklisted ID.
var ErrBlacklistedID = &errors.ErrDescriptor{
	MessageFormat:  "`{id}` is not an allowed ID",
	Code:           1,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"id"},
}

// ErrEmailAddressNotAllowed is returned when trying to set in an user account
// an email address that is not allowed.
var ErrEmailAddressNotAllowed = &errors.ErrDescriptor{
	MessageFormat:  "Email address is not allowed. Current allowed email addresses by the network administrators are in the following form: {allowed_emails}",
	Code:           2,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"allowed_emails"},
}

// ErrInvalidPassword is returned when trying to update the user's account
// password while providing a wrong current password.
var ErrInvalidPassword = &errors.ErrDescriptor{
	MessageFormat: "Invalid password",
	Code:          3,
	Type:          errors.InvalidArgument,
}

// ErrNotAuthorized is returned when trying to access a protected resource without
// authorization.
var ErrNotAuthorized = &errors.ErrDescriptor{
	MessageFormat: "Not authorized",
	Code:          4,
	Type:          errors.Unauthorized,
}

// ErrSetApplicationCollaboratorFailed is returned when after modifying an
// application's collaborators the application reaches an unmanageable state
// as the sum of rights that all collaborators with `RIGHT_APPLICATION_SETTINGS_COLLABORATORS`
// right is not equal to the entire set of available `RIGHT_APPLICATION_XXXXXX` rights.
var ErrSetApplicationCollaboratorFailed = &errors.ErrDescriptor{
	MessageFormat:  "Failed to set collaborator: application becomes unmanageable as no collaborator with `RIGHT_APPLICATION_SETTINGS_COLLABORATORS` has the following rights: {missing_rights}",
	Code:           5,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"missing_rights"},
}

// ErrSetGatewayCollaboratorFailed is returned when after modifying a gateway's
// collaborators the gateway reaches an unmanageable state as the sum of
// rights that all collaborators with `RIGHT_GATEWAY_SETTINGS_COLLABORATORS`
// right is not equal to the entire set of available `RIGHT_GATEWAY_XXXXXX` rights.
var ErrSetGatewayCollaboratorFailed = &errors.ErrDescriptor{
	MessageFormat:  "Failed to set collaborator: gateway becomes unmanageable as no collaborator with `RIGHT_GATEWAY_SETTINGS_COLLABORATORS` has the following rights: {missing_rights}",
	Code:           6,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"missing_rights"},
}

// ErrEmailAlreadyValidated is returned when calling 'RequestUserEmailValidation'
// when the email is already validated.
var ErrEmailAlreadyValidated = &errors.ErrDescriptor{
	MessageFormat: "Email address is already validated",
	Code:          7,
	Type:          errors.InvalidArgument,
}

// ErrValidationTokenExpired is returned when trying to validate an email address
// using an expired token.
var ErrValidationTokenExpired = &errors.ErrDescriptor{
	MessageFormat: "Token is expired",
	Code:          8,
	Type:          errors.PermissionDenied,
}

// ErrInvitationTokenMissing is returned when making a call to `CreateUser` with
// an empty invitation token and the self user registration is disabled.
var ErrInvitationTokenMissing = &errors.ErrDescriptor{
	MessageFormat: "Self account registration is disabled: a valid invitation token must be provided",
	Code:          9,
	Type:          errors.InvalidArgument,
}

// ErrEmailAddressAlreadyUsed is returned when an admin is trying to send an
// invitation to an email that is already being used by an user.
var ErrEmailAddressAlreadyUsed = &errors.ErrDescriptor{
	MessageFormat: "Failed to issue invitation: the email address is already in use",
	Code:          10,
	Type:          errors.InvalidArgument,
}

// ErrSetOrganizationMemberFailed is returned when after modifying an organization's
// members the organization reaches an unmanageable state as the sum of rights
// that all members with `RIGHT_ORGANIZATION_SETTINGS_MEMBERS` right is not equal
// to the entire set of available `RIGHT_ORGANIZATION_XXXXXX` rights.
var ErrSetOrganizationMemberFailed = &errors.ErrDescriptor{
	MessageFormat:  "Failed to set member: organization becomes unmanageable as no member with `RIGHT_ORGANIZATION_SETTINGS_MEMBERS` has the following rights: {missing_rights}",
	Code:           11,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"missing_rights"},
}

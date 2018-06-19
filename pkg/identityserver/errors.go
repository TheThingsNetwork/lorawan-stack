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

package identityserver

import (
	"go.thethings.network/lorawan-stack/pkg/errors"
	errorsv3 "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

func init() {
	ErrBlacklistedID.Register()
	ErrEmailAddressNotAllowed.Register()
	ErrInvalidPassword.Register()
	ErrEmailAlreadyValidated.Register()
	ErrValidationTokenExpired.Register()
	ErrInvitationTokenMissing.Register()
	ErrEmailAddressAlreadyUsed.Register()
	ErrUnmanageableApplication.Register()
	ErrUnmanageableGateway.Register()
	ErrUnmanageableOrganization.Register()
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

// ErrEmailAlreadyValidated is returned when calling 'RequestUserEmailValidation'
// when the email is already validated.
var ErrEmailAlreadyValidated = &errors.ErrDescriptor{
	MessageFormat: "Email address is already validated",
	Code:          4,
	Type:          errors.InvalidArgument,
}

// ErrValidationTokenExpired is returned when trying to validate an email address
// using an expired token.
var ErrValidationTokenExpired = &errors.ErrDescriptor{
	MessageFormat: "Token is expired",
	Code:          5,
	Type:          errors.PermissionDenied,
}

// ErrInvitationTokenMissing is returned when making a call to `CreateUser` with
// an empty invitation token and the self user registration is disabled.
var ErrInvitationTokenMissing = &errors.ErrDescriptor{
	MessageFormat: "Self account registration is disabled: a valid invitation token must be provided",
	Code:          6,
	Type:          errors.InvalidArgument,
}

// ErrEmailAddressAlreadyUsed is returned when an admin is trying to send an
// invitation to an email that is already being used by an user.
var ErrEmailAddressAlreadyUsed = &errors.ErrDescriptor{
	MessageFormat: "Failed to issue invitation: the email address is already in use",
	Code:          7,
	Type:          errors.InvalidArgument,
}

// ErrUnmanageableApplication is returned when the sum of rights of collaborators
// with `SETTINGS_COLLABORATORS` right is not equal to the entire set of defined
// application rights.
var ErrUnmanageableApplication = &errors.ErrDescriptor{
	MessageFormat:  "Application `{application_id}` becomes unmanageable as no collaborator with `RIGHT_APPLICATION_SETTINGS_COLLABORATORS` has the following rights: {missing_rights}",
	Code:           8,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"application_id", "missing_rights"},
}

// ErrUnmanageableGateway is returned when the sum of rights of collaborators
// with `SETTINGS_COLLABORATORS` right is not equal to the entire set of defined
// gateway rights.
var ErrUnmanageableGateway = &errors.ErrDescriptor{
	MessageFormat:  "Gateway `{gateway_id}` becomes unmanageable as no collaborator with `RIGHT_GATEWAY_SETTINGS_COLLABORATORS` has the following rights: {missing_rights}",
	Code:           9,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"gateway_id", "missing_rights"},
}

// ErrUnmanageableOrganization is returned when the sum of rights of collaborators
// with `SETTINGS_MEMBERS` right is not equal to the entire set of defined
// organization rights.
var ErrUnmanageableOrganization = &errors.ErrDescriptor{
	MessageFormat:  "Organization `{organization_id}` becomes unmanageable as no member with `RIGHT_ORGANIZATION_SETTINGS_MEMBERS` has the following rights: {missing_rights}",
	Code:           10,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"organization_id", "missing_rights"},
}

var (
	errOtherPullConfigurationStreamOpened = errorsv3.DefineAborted("other_pull_configuration_stream_opened", "another `PullConfiguration` stream opened")

	errWrongGatewayForAPIKey = errorsv3.DefinePermissionDenied("wrong_gateway_for_api_key", "api key has no rights for gateway `{gateway_id}`")
)

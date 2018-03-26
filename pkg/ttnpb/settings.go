// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "regexp"

var (
	// FieldPathSettingsBlacklistedIDs is the field path for the blacklisted IDs field.
	FieldPathSettingsBlacklistedIDs = regexp.MustCompile(`^blacklisted_ids$`)

	// FieldPathSettingsUserRegistrationSkipValidation is the field path for the
	// user registration flow skip validation field.
	FieldPathSettingsUserRegistrationSkipValidation = regexp.MustCompile(`^user_registration.skip_validation$`)

	// FieldPathSettingsUserRegistrationInvitationOnly is the field path for the
	// user registration flow invitation only field.
	FieldPathSettingsUserRegistrationInvitationOnly = regexp.MustCompile(`^user_registration.invitation_only$`)

	// FieldPathSettingsUserRegistrationAdminApproval is the field path for the
	// user registration flow admin approval field.
	FieldPathSettingsUserRegistrationAdminApproval = regexp.MustCompile(`^user_registration.admin_approval$`)

	// FieldPathSettingsValidationTokenTTL is the field path for the validation token TTL field.
	FieldPathSettingsValidationTokenTTL = regexp.MustCompile(`^validation_token_ttl$`)

	// FieldPathSettingsAllowedEmails is the field path for the allowed emails field.
	FieldPathSettingsAllowedEmails = regexp.MustCompile(`^allowed_emails$`)

	// FieldPathSettingsInvitationTokenTTL is the field path for the invitation token TTL field.
	FieldPathSettingsInvitationTokenTTL = regexp.MustCompile(`^invitation_token_ttl$`)
)

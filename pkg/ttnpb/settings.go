// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "regexp"

var (
	// FieldPathSettingsBlacklistedIDs is the field path for the blacklisted IDs field.
	FieldPathSettingsBlacklistedIDs = regexp.MustCompile(`^blacklisted_ids$`)

	// FieldPathSettingsAutomaticApproval is the field path for the automatic approval field.
	FieldPathSettingsAutomaticApproval = regexp.MustCompile(`^automatic_approval$`)

	// FieldPathSettingsClosedRegistration is the field path for the closed registration field.
	FieldPathSettingsClosedRegistration = regexp.MustCompile(`^closed_registration$`)

	// FieldPathSettingsValidationTokenTTL is the field path for the validation token TTL field.
	FieldPathSettingsValidationTokenTTL = regexp.MustCompile(`^validation_token_ttl$`)

	// FieldPathSettingsAllowedEmails is the field path for the allowed emails field.
	FieldPathSettingsAllowedEmails = regexp.MustCompile(`^allowed_emails$`)
)

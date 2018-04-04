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

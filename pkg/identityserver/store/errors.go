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

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

// Errors returned when stores can't find an entity.
var (
	ErrEntityNotFound = errors.DefineNotFound(
		"entity_not_found", "{entity_type} entity with id `{entity_id}` not found",
	)
	ErrAccountNotFound = errors.DefineNotFound(
		"account_not_found", "{account_type} account with id `{account_id}` not found",
	)
	ErrMembershipNotFound = errors.DefineNotFound(
		"membership_not_found", "membership of {account_type} account with id `{account_id}` on {entity_type} entity with id `{entity_id}` not found", //nolint:lll
	)

	ErrApplicationNotFound = errors.DefineNotFound(
		"application_not_found", "application with id `{application_id}` not found",
	)
	ErrClientNotFound = errors.DefineNotFound(
		"client_not_found", "client with id `{client_id}` not found",
	)
	ErrEndDeviceNotFound = errors.DefineNotFound(
		"end_device_not_found", "end device with id `{device_id}` not found in application with id `{application_id}`",
	)
	ErrGatewayNotFound = errors.DefineNotFound(
		"gateway_not_found", "gateway with id `{gateway_id}` not found",
	)
	ErrGatewayNotFoundByEUI = errors.DefineNotFound(
		"gateway_not_found_by_eui", "gateway with eui `{gateway_eui}` not found",
	)
	ErrOrganizationNotFound = errors.DefineNotFound(
		"organization_not_found", "organization with id `{organization_id}` not found",
	)
	ErrUserNotFound = errors.DefineNotFound(
		"user_not_found", "user with id `{user_id}` not found",
	)
	ErrUserNotFoundByPrimaryEmailAddress = errors.DefineNotFound(
		"user_not_found_by_primary_email_address", "user not found by primary email address",
	)

	ErrUserSessionNotFound = errors.DefineNotFound(
		"user_session_not_found", "user session with id `{session_id}` not found", "user_id",
	)
	ErrLastAdmin = errors.DefineFailedPrecondition(
		"last_admin", "user `{user_id}` is the last admin",
	)

	ErrAPIKeyNotFound = errors.DefineNotFound(
		"api_key_not_found", "api key with id `{api_key_id}` not found", "entity_type", "entity_id",
	)

	ErrInvitationAlreadySent = errors.DefineAlreadyExists(
		"invitation_already_sent", "invitation already sent",
	)
	ErrInvitationNotFound = errors.DefineNotFound(
		"invitation_not_found", "invitation not found", "invitation_token",
	)
	ErrInvitationExpired = errors.DefineFailedPrecondition(
		"invitation_expired", "invitation expired", "invitation_token",
	)
	ErrInvitationAlreadyUsed = errors.DefineFailedPrecondition(
		"invitation_already_used", "invitation already used", "invitation_token",
	)

	ErrValidationAlreadySent = errors.DefineAlreadyExists(
		"validation_already_sent", "validation already sent",
	)
	ErrValidationTokenNotFound = errors.DefineNotFound(
		"validation_token_not_found", "validation token not found", "validation_id",
	)
	ErrValidationTokenExpired = errors.DefineFailedPrecondition(
		"validation_token_expired", "validation token expired", "validation_id",
	)
	ErrValidationTokenAlreadyUsed = errors.DefineFailedPrecondition(
		"validation_token_already_used", "validation token already used", "validation_id",
	)
	ErrValidationWithoutContactInfo = errors.DefineNotFound(
		"validation_without_contact_info", "validation does not reference any valid contact information",
	)

	ErrLoginTokenNotFound = errors.DefineNotFound(
		"login_token_not_found", "login token not found",
	)
	ErrLoginTokenExpired = errors.DefineFailedPrecondition(
		"login_token_expired", "login token expired",
	)
	ErrLoginTokenAlreadyUsed = errors.DefineFailedPrecondition(
		"login_token_already_used", "login token already used",
	)

	ErrAuthorizationNotFound = errors.DefineNotFound(
		"authorization_not_found", "authorization of user with id `{user_id}` on client with id `{client_id}` not found",
	)
	ErrAuthorizationCodeNotFound = errors.DefineNotFound(
		"authorization_code_not_found", "authorization code not found",
	)
	ErrAccessTokenNotFound = errors.DefineNotFound(
		"access_token_not_found", "access token with id `{access_token_id}` not found",
	)

	ErrNoEUIBlockAvailable = errors.DefineFailedPrecondition(
		"no_eui_or_block_available",
		"no EUI or EUI block available",
	)
	ErrApplicationDevEUILimitReached = errors.DefineFailedPrecondition(
		"application_dev_eui_limit_reached",
		"application issued DevEUI limit ({dev_eui_limit}) reached",
	)
)

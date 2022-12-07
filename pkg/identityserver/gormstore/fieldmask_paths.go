// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

const (
	// NOTE: please keep this sorted.
	activatedAtField                    = "activated_at"
	adminField                          = "admin"
	administrativeContactField          = "administrative_contact"
	antennasField                       = "antennas"
	applicationServerAddressField       = "application_server_address"
	attributesField                     = "attributes"
	autoUpdateField                     = "auto_update"
	bandIDField                         = "version_ids.band_id"
	brandIDField                        = "version_ids.brand_id"
	claimAuthenticationCodeField        = "claim_authentication_code"
	contactInfoField                    = "contact_info"
	descriptionField                    = "description"
	devEuiCounterField                  = "dev_eui_counter"
	disablePacketBrokerForwardingField  = "disable_packet_broker_forwarding"
	downlinkPathConstraintField         = "downlink_path_constraint"
	endorsedField                       = "endorsed"
	enforceDutyCycleField               = "enforce_duty_cycle"
	firmwareVersionField                = "version_ids.firmware_version"
	frequencyPlanIDsField               = "frequency_plan_ids"
	gatewayServerAddressField           = "gateway_server_address"
	grantsField                         = "grants"
	hardwareVersionField                = "version_ids.hardware_version"
	joinServerAddressField              = "join_server_address"
	lastSeenAtField                     = "last_seen_at"
	lbsLNSSecretField                   = "lbs_lns_secret" //nolint:gosec
	locationPublicField                 = "location_public"
	locationsField                      = "locations"
	logoutRedirectURIsField             = "logout_redirect_uris"
	lrfhssField                         = "lrfhss"
	lrfhssSupportedField                = "lrfhss.supported"
	modelIDField                        = "version_ids.model_id"
	nameField                           = "name"
	networkServerAddressField           = "network_server_address"
	passwordField                       = "password"
	passwordUpdatedAtField              = "password_updated_at"
	pictureField                        = "picture"
	primaryEmailAddressField            = "primary_email_address"
	primaryEmailAddressValidatedAtField = "primary_email_address_validated_at"
	profilePictureField                 = "profile_picture"
	redirectURIsField                   = "redirect_uris"
	requireAuthenticatedConnectionField = "require_authenticated_connection"
	requirePasswordUpdateField          = "require_password_update"
	rightsField                         = "rights"
	scheduleAnytimeDelayField           = "schedule_anytime_delay"
	scheduleDownlinkLateField           = "schedule_downlink_late"
	secretField                         = "secret"
	serialNumberField                   = "serial_number"
	serviceProfileIDField               = "service_profile_id"
	skipAuthorizationField              = "skip_authorization"
	stateDescriptionField               = "state_description"
	stateField                          = "state"
	statusPublicField                   = "status_public"
	targetCUPSKeyField                  = "target_cups_key"
	targetCUPSURIField                  = "target_cups_uri"
	technicalContactField               = "technical_contact"
	temporaryPasswordCreatedAtField     = "temporary_password_created_at"
	temporaryPasswordExpiresAtField     = "temporary_password_expires_at"
	temporaryPasswordField              = "temporary_password"
	updateChannelField                  = "update_channel"
	updateLocationFromStatusField       = "update_location_from_status"
	versionIDsField                     = "version_ids"
)
